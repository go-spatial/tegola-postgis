package register

import (
	"fmt"
	"html"

	"github.com/go-spatial/geom"
	tegola "github.com/go-spatial/tegola-postgis"
	"github.com/go-spatial/tegola-postgis/atlas"
	"github.com/go-spatial/tegola-postgis/config"
	"github.com/go-spatial/tegola-postgis/mvtprovider"
	"github.com/go-spatial/tegola-postgis/provider"
)

type ErrProviderLayerInvalid struct {
	ProviderLayer string
	Map           string
}

func (e ErrProviderLayerInvalid) Error() string {
	return fmt.Sprintf("invalid provider layer (%v) for map (%v)", e.ProviderLayer, e.Map)
}

type ErrProviderNotFound struct {
	Provider string
}

func (e ErrProviderNotFound) Error() string {
	return fmt.Sprintf("provider (%v) not defined", e.Provider)
}

type ErrProviderLayerNotRegistered struct {
	MapName       string
	ProviderLayer string
	Provider      string
}

func (e ErrProviderLayerNotRegistered) Error() string {
	return fmt.Sprintf("map (%v) 'provider_layer' (%v) is not registered with provider (%v)", e.MapName, e.ProviderLayer, e.Provider)
}

type ErrFetchingLayerInfo struct {
	Provider string
}

func (e ErrFetchingLayerInfo) Error() string {
	return fmt.Sprintf("error fetching layer info from provider (%v)", e.Provider)
}

type ErrDefaultTagsInvalid struct {
	ProviderLayer string
}

func (e ErrDefaultTagsInvalid) Error() string {
	return fmt.Sprintf("'default_tags' for 'provider_layer' (%v) should be a TOML table", e.ProviderLayer)
}

func initLayer(l *config.MapLayer, mname string, lprovider provider.Layerer) (atlas.Layer, error) {
	// read the provider's layer names
	pname, lname, _ := l.GetProviderLayerName()
	layerInfos, err := lprovider.Layers()
	if err != nil {
		return atlas.Layer{}, ErrFetchingLayerInfo{
			Provider: pname,
		}
	}
	providerLayer := string(l.ProviderLayer)

	// confirm our providerLayer name is registered
	var found bool
	var layerGeomType geom.Geometry
	for i := range layerInfos {
		if layerInfos[i].Name() == lname {
			found = true

			// read the layerGeomType
			layerGeomType = layerInfos[i].GeomType()
			break
		}
	}
	if !found {
		return atlas.Layer{}, ErrProviderLayerNotRegistered{
			MapName:       mname,
			ProviderLayer: providerLayer,
			Provider:      pname,
		}
	}

	var defaultTags map[string]interface{}
	if l.DefaultTags != nil {
		var ok bool
		defaultTags, ok = l.DefaultTags.(map[string]interface{})
		if !ok {
			return atlas.Layer{}, ErrDefaultTagsInvalid{
				ProviderLayer: providerLayer,
			}
		}
	}

	var minZoom uint
	if l.MinZoom != nil {
		minZoom = uint(*l.MinZoom)
	}

	var maxZoom uint
	if l.MaxZoom != nil {
		maxZoom = uint(*l.MaxZoom)
	}

	prvd, _ := lprovider.(provider.Tiler)

	// add our layer to our layers slice
	return atlas.Layer{
		Name:              string(l.Name),
		ProviderLayerName: lname,
		MinZoom:           minZoom,
		MaxZoom:           maxZoom,
		Provider:          prvd,
		DefaultTags:       defaultTags,
		GeomType:          layerGeomType,
		DontSimplify:      bool(l.DontSimplify),
		DontClip:          bool(l.DontClip),
	}, nil
}

// Maps registers maps with with atlas
func Maps(a *atlas.Atlas, maps []config.Map, providers map[string]provider.Tiler, mvtProviders map[string]mvtprovider.Tiler) error {

	// iterate our maps
	for _, m := range maps {
		newMap := atlas.NewWebMercatorMap(string(m.Name))
		newMap.Attribution = html.EscapeString(string(m.Attribution))

		// convert from env package
		centerArr := [3]float64{}
		for i, v := range m.Center {
			centerArr[i] = float64(v)
		}

		newMap.Center = centerArr

		if len(m.Bounds) == 4 {
			newMap.Bounds = geom.NewExtent(
				[2]float64{float64(m.Bounds[0]), float64(m.Bounds[1])},
				[2]float64{float64(m.Bounds[2]), float64(m.Bounds[3])},
			)
		}

		if m.TileBuffer == nil {
			newMap.TileBuffer = tegola.DefaultTileBuffer
		} else {
			newMap.TileBuffer = uint64(*m.TileBuffer)
		}

		// iterate our layers
		for _, l := range m.Layers {
			// split our provider name (provider.layer) into [provider,layer]
			providerName, _, err := l.GetProviderLayerName()
			if err != nil {
				return ErrProviderLayerInvalid{
					ProviderLayer: string(l.ProviderLayer),
					Map:           string(m.Name),
				}
			}

			var layerer provider.Layerer

			// lookup our provider

			// first check to see if we are an mvt type of map.
			if newMap.HasMVTProvider() {
				// okay we need to make sure the provider names match up.
				if newMap.MVTProviderName() != providerName {
					return config.ErrMVTDiffereProviders{
						Original: newMap.MVTProviderName(),
						Current:  providerName,
					}
				}
				layerer = newMap.MVTProvider()
				goto ADDLAYER
			}

			// search for it in the normal providers
			if prvd, ok := providers[providerName]; ok {
				layerer = prvd
				goto ADDLAYER
			}

			// search for it in the mvt providers
			if mvtprvd, ok := mvtProviders[providerName]; ok {
				layerer = newMap.SetMVTProvider(providerName, mvtprvd)
				goto ADDLAYER
			}

			// We did not find the provider
			return ErrProviderNotFound{providerName}

		ADDLAYER:

			newLayer, err := initLayer(&l, string(m.Name), layerer)
			if err != nil {
				return err
			}
			newMap.Layers = append(newMap.Layers, newLayer)
		}
		a.AddMap(newMap)
	}

	return nil
}

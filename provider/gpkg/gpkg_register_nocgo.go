// +build !cgo

package gpkg

import "github.com/go-spatial/tegola-postgis/provider"

func NewTileProvider(config map[string]interface{}) (provider.Tiler, error) {
	return nil, provider.ErrUnsupported
}

func Cleanup() {}

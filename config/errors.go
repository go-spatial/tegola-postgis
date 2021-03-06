package config

import "fmt"

type ErrMapNotFound struct {
	MapName string
}

func (e ErrMapNotFound) Error() string {
	return fmt.Sprintf("config: map (%v) not found", e.MapName)
}

type ErrInvalidProviderLayerName struct {
	ProviderLayerName string
}

func (e ErrInvalidProviderLayerName) Error() string {
	return fmt.Sprintf("config: invalid provider layer name (%v)", e.ProviderLayerName)
}

type ErrOverlappingLayerZooms struct {
	ProviderLayer1 string
	ProviderLayer2 string
}

func (e ErrOverlappingLayerZooms) Error() string {
	return fmt.Sprintf("config: overlapping zooms for layer (%v) and layer (%v)", e.ProviderLayer1, e.ProviderLayer2)
}

type ErrMVTDiffereProviders struct {
	Original string
	Current  string
}

func (e ErrMVTDiffereProviders) Error() string {
	return fmt.Sprintf(
		"config: all layer providers need to be the same, first provider is %s second provider is %s",
		e.Original,
		e.Current,
	)
}

type ErrMixedProviders struct {
	Map string
}

func (e ErrMixedProviders) Error() string {
	return fmt.Sprintf("config: can not mix MVT providers with normal providers for map %v", e.Map)
}

type ErrInvalidLayerZoom struct {
	ProviderLayer string
	MinZoom       bool
	Zoom          int
	ZoomLimit     int
}

func (e ErrInvalidLayerZoom) Error() string {
	n, d := "MaxZoom", "above"
	if e.MinZoom {
		n, d = "MinZoom", "below"
	}
	return fmt.Sprintf(
		"config: for provider layer %v %v(%v) is %v allowed level of %v",
		e.ProviderLayer, n, e.Zoom, d, e.ZoomLimit,
	)
}

type ErrMissingEnvVar struct {
	EnvVar string
}

func (e ErrMissingEnvVar) Error() string {
	return fmt.Sprintf("config: config file is referencing an environment variable that is not set (%v)", e.EnvVar)
}

type ErrInvalidHeader struct {
	Header string
}

func (e ErrInvalidHeader) Error() string {
	return fmt.Sprintf("config: header (%v) blacklisted", e.Header)
}

type ErrInvalidURIPrefix string

func (e ErrInvalidURIPrefix) Error() string {
	return fmt.Sprintf("config: invalid uri_prefix (%v). uri_prefix must start with a forward slash '/' ", string(e))
}

package atlas

// The point of this file is to load and register the default cache backends
import (
	_ "github.com/go-spatial/tegola-postgis/cache/file"
)

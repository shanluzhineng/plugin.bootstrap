package options

const (
	ConfigurationKey string = "plugininstaller"
)

type Options struct {
	BuildingPlugIns []PlugInMetabaseConfiguration `json:"buildingPlugIns"`
}

type PlugInMetabaseConfiguration struct {
	Name     string `json:"name"`
	FeedName string `json:"feedName"`
	Path     string `json:"path"`
}

package config

// KeymapConfig represents key mappings
type KeymapConfig struct {
	Normal map[string]interface{} `toml:"normal"`
	Insert map[string]interface{} `toml:"insert"`
}

package config

type LauncherConfig struct {
	Nick     string `toml:"nick"`
	Version  string `toml:"version"`
	Instance string `toml:"instance"`
}

type InstanceConfig struct {
	ID     string `toml:"id"`
	Name   string `toml:"name"` // Instance name
	Branch string `toml:"branch"`
	Build  int    `toml:"build"` // Game build aka version
}

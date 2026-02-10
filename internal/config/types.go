package config

type LauncherConfig struct {
	Nick       string `toml:"nick"`
	Instance   string `toml:"instance"`
	DiscordRPC bool   `toml:"discord_rpc"`
}

type InstanceConfig struct {
	ID     string `toml:"id"`
	Name   string `toml:"name"`
	Branch string `toml:"branch"`
	Build  string `toml:"build"`
}

package config

type LauncherConfig struct {
	Nick       string `toml:"nick"`
	Version    string `toml:"version"`
	Instance   string `toml:"instance"`
	DiscordRPC bool   `toml:"discord_rpc"`
	GameDir    string `toml:"game_dir,omitempty"`
}

type InstanceConfig struct {
	ID     string `toml:"id"`
	Name   string `toml:"name"`
	Branch string `toml:"branch"`
	Build  string `toml:"build"`
}

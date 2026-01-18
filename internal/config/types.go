package config

type Profile struct {
	ID   string `toml:"id" json:"id"`
	Name string `toml:"name" json:"name"`
}

type GameSettings struct {
	MinMemory   uint   `toml:"min_memory" json:"minMemory"`
	MaxMemory   uint   `toml:"max_memory" json:"maxMemory"`
	Width       int    `toml:"width" json:"width"`
	Height      int    `toml:"height" json:"height"`
	Fullscreen  bool   `toml:"fullscreen" json:"fullscreen"`
	JavaArgs    string `toml:"java_args" json:"javaArgs"`
	GameDir     string `toml:"game_dir" json:"gameDir"`
	Channel     string `toml:"channel" json:"channel"`
	GameVersion int    `toml:"game_version" json:"gameVersion"`
	OnlineFix   bool   `toml:"online_fix" json:"onlineFix"`
}

type Config struct {
	Version        string       `toml:"version" json:"version"`
	Profiles       []Profile    `toml:"profiles" json:"profiles"`
	CurrentProfile string       `toml:"current_profile" json:"current_profile"`
	Settings       GameSettings `toml:"settings" json:"settings"`
}

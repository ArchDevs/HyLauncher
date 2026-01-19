package config

type Config struct {
	Version            string `toml:"version"`
	Nick               string `toml:"nick"`
	CurrentGameVersion int    `toml:"current_game_version"`
	Branch             string `toml:"branch"`
}

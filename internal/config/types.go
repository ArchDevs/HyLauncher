package config

type Profile struct {
	ID   string `toml:"id" json:"id"`
	Name string `toml:"name" json:"name"`
}

type Config struct {
	Version        string    `toml:"version"`
	Profiles       []Profile `toml:"profiles"`
	CurrentProfile string    `toml:"current_profile"`
	Nick           string    `toml:"nick"` // Kept for backward compatibility or migration
}

package config

type Profile struct {
	ID   string `toml:"id" json:"id"`
	Name string `toml:"name" json:"name"`
}

type Config struct {
	Version        string    `toml:"version" json:"version"`
	Profiles       []Profile `toml:"profiles" json:"profiles"`
	CurrentProfile string    `toml:"current_profile" json:"current_profile"`
}


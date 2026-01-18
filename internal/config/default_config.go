package config

import "github.com/google/uuid"

func Default() Config {
	id := uuid.New().String()
	return Config{
		Version: "0.6.5",
		Profiles: []Profile{
			{
				ID:   id,
				Name: "HyLauncher",
			},
		},
		CurrentProfile: id,
	}
}

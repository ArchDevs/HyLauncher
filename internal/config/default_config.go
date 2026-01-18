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
		Settings: GameSettings{
			MinMemory:  2,
			MaxMemory:  4,
			Width:      1024,
			Height:     640,
			Fullscreen: false,
			JavaArgs:   "-XX:+UseG1GC -Dsun.rmi.dgc.server.gcInterval=2147483646 -XX:+UnlockExperimentalVMOptions -XX:G1NewSizePercent=20 -XX:G1ReservePercent=20 -XX:MaxGCPauseMillis=50 -XX:G1HeapRegionSize=32M",
			GameDir:    "",
		},
	}
}

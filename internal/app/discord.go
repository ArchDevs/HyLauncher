package app

import (
	"fmt"
	"time"

	"HyLauncher/internal/config"

	"github.com/hugolgst/rich-go/client"
)

func (a *App) discordRPC() {
	if !a.launcherCfg.DiscordRPC {
		return
	}

	now := time.Now()

	err := client.SetActivity(client.Activity{
		State:   "Idle",
		Details: "The best Hytale launcher",
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{
			{
				Label: "GitHub",
				Url:   "https://github.com/ArchDevs/HyLauncher",
			},
			{
				Label: "Website",
				Url:   "https://hylauncher.fun",
			},
		},
	})

	_ = err
}

func (a *App) GetDiscordRPC() bool {
	return a.launcherCfg.DiscordRPC
}

func (a *App) SetDiscordRPC(enabled bool) error {
	// Update config
	if err := config.UpdateLauncher(func(cfg *config.LauncherConfig) error {
		cfg.DiscordRPC = enabled
		return nil
	}); err != nil {
		return fmt.Errorf("failed to save Discord RPC setting: %w", err)
	}

	a.launcherCfg.DiscordRPC = enabled

	if enabled {
		if err := client.Login("1465005878276128888"); err != nil {
			return nil
		}
		go a.discordRPC()
	} else {
		client.Logout()
	}

	return nil
}

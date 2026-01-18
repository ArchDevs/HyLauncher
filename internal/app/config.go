package app

import (
	"HyLauncher/internal/config"
	"fmt"

	"github.com/google/uuid"
)

func (a *App) GetProfiles() []config.Profile {
	return a.cfg.Profiles
}

func (a *App) GetCurrentProfile() config.Profile {
	for _, p := range a.cfg.Profiles {
		if p.ID == a.cfg.CurrentProfile {
			return p
		}
	}
	if len(a.cfg.Profiles) > 0 {
		return a.cfg.Profiles[0]
	}
	return config.Profile{}
}

func (a *App) SetCurrentProfile(id string) error {
	a.cfg.CurrentProfile = id
	return config.Save(a.cfg)
}

func (a *App) AddProfile(name string) (config.Profile, error) {
	newProfile := config.Profile{
		ID:   uuid.New().String(),
		Name: name,
	}
	a.cfg.Profiles = append(a.cfg.Profiles, newProfile)
	a.cfg.CurrentProfile = newProfile.ID
	err := config.Save(a.cfg)
	return newProfile, err
}

func (a *App) UpdateProfile(id string, name string) error {
	for i, p := range a.cfg.Profiles {
		if p.ID == id {
			a.cfg.Profiles[i].Name = name
			return config.Save(a.cfg)
		}
	}
	return fmt.Errorf("profile not found")
}

func (a *App) DeleteProfile(id string) error {
	if len(a.cfg.Profiles) <= 1 {
		return fmt.Errorf("cannot delete last profile")
	}

	index := -1
	for i, p := range a.cfg.Profiles {
		if p.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("profile not found")
	}

	a.cfg.Profiles = append(a.cfg.Profiles[:index], a.cfg.Profiles[index+1:]...)

	if a.cfg.CurrentProfile == id {
		a.cfg.CurrentProfile = a.cfg.Profiles[0].ID
	}

	return config.Save(a.cfg)
}

func (a *App) SetNick(nick string) error {
	// Update name of current profile for compatibility
	for i, p := range a.cfg.Profiles {
		if p.ID == a.cfg.CurrentProfile {
			a.cfg.Profiles[i].Name = nick
			return config.Save(a.cfg)
		}
	}
	return nil
}

func (a *App) GetNick() string {
	return a.GetCurrentProfile().Name
}

func (a *App) GetLauncherVersion() string {
	return config.Default().Version
}

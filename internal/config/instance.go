package config

import (
	"HyLauncher/internal/env"
	"path/filepath"
)

func instancePath(instanceID string) string {
	return filepath.Join(env.GetInstanceDir(instanceID), "config.toml")
}

func LoadInstance(instanceID string) (*InstanceConfig, error) {
	return load(instancePath(instanceID), InstanceDefault)
}

func SaveInstance(instanceID string, cfg *InstanceConfig) error {
	return save(instancePath(instanceID), cfg)
}

func UpdateInstance(id string, update func(*InstanceConfig) error) error {
	cfg, err := LoadInstance(id)
	if err != nil {
		return err
	}

	if err := update(cfg); err != nil {
		return err
	}

	return SaveInstance(id, cfg)
}

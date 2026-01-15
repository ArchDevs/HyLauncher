package config

import (
	"HyLauncher/internal/env"
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Nick string `json:"nick"`
}

func getConfigPath() string {
	dir := env.GetDefaultAppDir()
	return filepath.Join(dir, "config.json")
}

func Save(cfg *Config) error {
	path := getConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(cfg)
}

func Load() (*Config, error) {
	path := getConfigPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	defer f.Close()

	cfg := &Config{}
	err = json.NewDecoder(f).Decode(cfg)
	return cfg, err
}

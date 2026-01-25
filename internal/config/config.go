package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func load[T any](path string, defaults func() T) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaults()
			_ = save(path, &cfg)
			return &cfg, nil
		}
		return nil, err
	}

	cfg := defaults()
	if err := toml.Unmarshal(data, &cfg); err != nil {
		_ = os.Rename(path, path+".broken")
		cfg = defaults()
		_ = save(path, &cfg)
		return &cfg, nil
	}

	return &cfg, nil
}

func save[T any](path string, cfg *T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

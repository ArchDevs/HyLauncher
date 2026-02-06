package service

import (
	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/model"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type InstanceService struct{}

func NewInstanceService() *InstanceService {
	return &InstanceService{}
}

func (s *InstanceService) CreateInstance(request model.InstanceModel) (*model.InstanceModel, error) {
	instanceID := makeInstanceID(request.InstanceName)

	instanceDir := env.GetInstanceDir(request.InstanceID)

	_ = os.MkdirAll(instanceDir, 0755)

	userDataDir := filepath.Join(instanceDir, "UserData")
	if ok := fileutil.FileExists(userDataDir); ok == false {
		_ = os.MkdirAll(userDataDir, 0755)
	}

	cfg := config.InstanceDefault()
	cfg.ID = instanceID
	cfg.Name = request.InstanceName
	cfg.Branch = request.Branch
	cfg.Build = request.BuildVersion

	return &model.InstanceModel{
		InstanceID:   cfg.ID,
		InstanceName: cfg.Name,
		Branch:       cfg.Branch,
		BuildVersion: cfg.Build,
	}, nil
}

func (s *InstanceService) DeleteInstance(instanceID string) {
	if _, err := os.Stat(env.GetInstanceDir(instanceID)); err != nil {
		return
	}
	_ = os.RemoveAll(instanceID)
}

func (s *InstanceService) ListInstances() ([]model.InstanceModel, error) {
	root := env.GetInstancesDir()

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var instances []model.InstanceModel

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		cfgPath := filepath.Join(root, entry.Name(), "config.toml")
		if !fileutil.FileExists(cfgPath) {
			continue
		}

		cfg, err := config.LoadInstance(cfgPath)
		if err != nil {
			continue
		}

		instances = append(instances, model.InstanceModel{
			InstanceID:   cfg.ID,
			InstanceName: cfg.Name,
			Branch:       cfg.Branch,
			BuildVersion: cfg.Build,
		})
	}

	return instances, nil
}

func makeInstanceID(name string) string {
	base := slug.Make(name)
	if base == "" {
		base = "instance"
	}

	shortID := uuid.New().String()[:6]
	return base + "-" + shortID
}

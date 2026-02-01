package service

import (
	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/model"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type InstanceService struct{}

func NewInstanceService() *InstanceService {
	return &InstanceService{}
}

func (s *InstanceService) CreateInstance(request model.InstanceModel) (*model.InstanceModel, error) {
	instanceID := makeInstanceSlug(request.InstanceName)

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

func (s *InstanceService) DeleteInstance() {}

func makeInstanceSlug(name string) string {
	base := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s_%d", base, rand.Intn(999999))
}

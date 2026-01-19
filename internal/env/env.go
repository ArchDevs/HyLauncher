package env

import (
	"os"
	"path/filepath"
	"runtime"
)

func GetOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return "unknown"
	}
}

func GetArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return "unknown"
	}
}

func GetDefaultAppDir() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local", "HyLauncher")
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "HyLauncher")
	case "linux":
		return filepath.Join(home, ".hylauncher")
	default:
		return filepath.Join(home, "HyLauncher")
	}
}

func GetCacheDir() string {
	return filepath.Join(GetDefaultAppDir(), "cache")
}

func CreateFolders(branch string) error {
	basePath := GetDefaultAppDir()
	packagePath := filepath.Join(basePath, branch, "package") // Package folder

	paths := []string{
		basePath,                                   // main folder
		filepath.Join(basePath, "UserData"),        // UserData
		filepath.Join(basePath, "cache"),           // cache Folder
		filepath.Join(basePath, "tools", "butler"), // Butler
		filepath.Join(packagePath, "jre"),          // JRE Folder
		filepath.Join(packagePath, "game"),
		filepath.Join(packagePath, "game", "latest"),
	}

	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	return nil
}

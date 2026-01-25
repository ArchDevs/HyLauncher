package env

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

func GetInstancesDir() string {
	return filepath.Join(GetDefaultAppDir(), "instances")
}

func GetServersDir() string {
	return filepath.Join(GetDefaultAppDir(), "servers")
}

func GetInstanceDir(instance string) string {
	return filepath.Join(GetInstancesDir(), instance)
}

// Deprecated, for backward co
func GetInstance(instance string) string {
	return GetInstanceDir(instance)
}

func GetInstanceUserDataDir(instance string) string {
	return filepath.Join(GetInstanceDir(instance), "UserData")
}

func GetJREDir() string {
	return filepath.Join(GetDefaultAppDir(), "shared", "jre")
}

func GetSharedGamesDir() string {
	return filepath.Join(GetDefaultAppDir(), "shared", "games")
}

func GetGameDir(branch string, version int) string {
	return filepath.Join(GetSharedGamesDir(), branch, strconv.Itoa(version))
}

func GetGameClientPath(branch string, version int) string {
	gameDir := GetGameDir(branch, version)
	if runtime.GOOS == "darwin" {
		return filepath.Join(gameDir, "Client", "Hytale.app", "Contents", "MacOS", "HytaleClient")
	} else if runtime.GOOS == "windows" {
		return filepath.Join(gameDir, "Client", "HytaleClient.exe")
	}
	return filepath.Join(gameDir, "Client", "HytaleClient")
}

func CreateFolders(instance string) error {
	basePath := GetDefaultAppDir()

	paths := []string{
		basePath,                                       // Main folder
		filepath.Join(basePath, "cache"),               // Cache Folder
		filepath.Join(basePath, "instances"),           // Game instances folder
		filepath.Join(basePath, "instances", instance), // Specific instance
		GetInstanceUserDataDir(instance),               // Instance UserData
		filepath.Join(basePath, "servers"),             // Servers folder
		filepath.Join(basePath, "logs"),                // Logs Folder
		filepath.Join(basePath, "crashes"),             // Crashes Folder
		filepath.Join(basePath, "shared"),              // Shared folder
		filepath.Join(basePath, "shared", "jre"),       // Shared JRE folder
		filepath.Join(basePath, "shared", "butler"),    // Butler
		filepath.Join(basePath, "shared", "games"),     // Shared games folder
	}

	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	return nil
}

package env

import (
	"os"
	"path/filepath"
	"runtime"
)

var customAppDir string

func SetCustomAppDir(path string) {
	customAppDir = path
}

func GetCustomAppDir() string {
	return customAppDir
}

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
	if customAppDir != "" {
		return customAppDir
	}

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

func GetInstanceUserDataDir(instance string) string {
	return filepath.Join(GetInstanceDir(instance), "UserData")
}

func GetJREDir() string {
	return filepath.Join(GetDefaultAppDir(), "shared", "jre")
}

func GetSharedGamesDir() string {
	return filepath.Join(GetDefaultAppDir(), "shared", "games")
}

func GetGameDir(branch string, version string) string {
	return filepath.Join(GetSharedGamesDir(), branch, version)
}

func GetServerPath(branch string, version string) string {
	gameDir := GetGameDir(branch, version)

	possible := []string{
		filepath.Join(gameDir, "Server", "HytaleServer.jar"),
		filepath.Join(gameDir, "Server", "server.jar"),
	}

	for _, p := range possible {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

func GetGameClientPath(branch string, version string) string {
	var possible []string
	gameDir := GetGameDir(branch, version)

	switch runtime.GOOS {
	case "darwin":
		possible = []string{
			filepath.Join(gameDir, "Client", "Hytale.app", "Contents", "MacOS", "HytaleClient"),
			filepath.Join(gameDir, "Client", "HytaleClient"),
		}
	case "windows":
		possible = []string{
			filepath.Join(gameDir, "Client", "HytaleClient.exe"),
		}
	default:
		possible = []string{
			filepath.Join(gameDir, "Client", "HytaleClient"),
		}
	}

	for _, p := range possible {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
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

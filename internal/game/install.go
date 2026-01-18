package game

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"HyLauncher/internal/env"
	"HyLauncher/internal/java"
	"HyLauncher/internal/patch"
)

var (
	installMutex sync.Mutex
	isInstalling bool
)

func EnsureInstalled(ctx context.Context, channel string, targetVersion int, enableOnlineFix bool, progress func(stage string, progress float64, msg string, file string, speed string, down, total int64)) error {
	// Prevent multiple simultaneous installations
	installMutex.Lock()
	if isInstalling {
		installMutex.Unlock()
		return fmt.Errorf("installation already in progress")
	}
	isInstalling = true
	installMutex.Unlock()

	defer func() {
		installMutex.Lock()
		isInstalling = false
		installMutex.Unlock()
	}()

	// Download JRE
	if err := java.DownloadJRE(ctx, progress); err != nil {
		return fmt.Errorf("failed to download Java Runtime: %w", err)
	}

	// Install Butler
	if _, err := patch.InstallButler(ctx, progress); err != nil {
		return fmt.Errorf("failed to install Butler tool: %w", err)
	}

	var latestVersion int
	var installDirName string
	var successURL string

	if targetVersion == 0 {
		installDirName = "latest"
		// Find latest version with details
		if progress != nil {
			progress("version", 0, "Checking for game updates...", "", "", 0, 0)
		}

		// Run version check (will use cache if available)
		result := patch.FindLatestVersionWithDetails(channel)

		if result.Error != nil {
			return fmt.Errorf(
				"cannot find game versions on server\n\n"+
					"Platform: %s %s\n"+
					"Error: %v\n\n"+
					"Troubleshooting:\n"+
					"• Ensure your system is supported (Windows/Linux/macOS)\n"+
					"• Check if game is available for your architecture\n"+
					"• Verify firewall allows connections to game-patches.hytale.com\n"+
					"• Try disabling VPN temporarily\n\n"+
					"Checked URLs: %d\n"+
					"Sample: %s",
				runtime.GOOS,
				runtime.GOARCH,
				result.Error,
				len(result.CheckedURLs),
				getFirstURL(result.CheckedURLs),
			)
		}

		if result.LatestVersion == 0 {
			return fmt.Errorf(
				"no game versions found for your platform\n\n"+
					"Platform: %s/%s\n"+
					"Version type: %s\n\n"+
					"This usually means:\n"+
					"• The game is not yet available for your platform\n"+
					"• Your system architecture is not supported\n"+
					"• Server configuration has changed\n\n"+
					"Please check the official Hytale website for platform availability.",
				runtime.GOOS,
				runtime.GOARCH,
				channel,
			)
		}
		latestVersion = result.LatestVersion
		successURL = result.SuccessURL
	} else {
		// Verify if the requested version exists
		if progress != nil {
			progress("version", 0, fmt.Sprintf("Verifying version %d...", targetVersion), "", "", 0, 0)
		}
		if err := patch.VerifyVersionExists(channel, targetVersion); err != nil {
			return fmt.Errorf("requested version %d is not available: %w", targetVersion, err)
		}
		latestVersion = targetVersion
		installDirName = strconv.Itoa(targetVersion)
	}

	if progress != nil {
		progress("version", 100, fmt.Sprintf("Found version %d", latestVersion), "", "", 0, 0)
	}

	fmt.Printf("Target version: %d\n", latestVersion)
	if successURL != "" {
		fmt.Printf("Success URL: %s\n", successURL)
	}

	return InstallGame(ctx, channel, latestVersion, installDirName, enableOnlineFix, progress)
}

func InstallGame(ctx context.Context, versionType string, remoteVer int, installDirName string, enableOnlineFix bool, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	localStr := patch.GetLocalVersion(versionType)
	local, _ := strconv.Atoi(localStr)

	gameInstallDir := filepath.Join(env.GetDefaultAppDir(), versionType, "package", "game", installDirName)

	gameClient := "HytaleClient"
	if runtime.GOOS == "windows" {
		gameClient += ".exe"
	}
	clientPath := filepath.Join(gameInstallDir, "Client", gameClient)
	_, clientErr := os.Stat(clientPath)

	if local == remoteVer && clientErr == nil {
		if progressCallback != nil {
			progressCallback("complete", 100, "Game is up to date", "", "", 0, 0)
		}
		return nil
	}

	prevVer := local
	if clientErr != nil {
		prevVer = 0
		if progressCallback != nil {
			progressCallback("download", 0, fmt.Sprintf("Installing game version %d...", remoteVer), "", "", 0, 0)
		}
	} else {
		if progressCallback != nil {
			progressCallback("download", 0, fmt.Sprintf("Updating from version %d to %d...", local, remoteVer), "", "", 0, 0)
		}
	}

	// Download the patch file
	pwrPath, err := patch.DownloadPWR(ctx, versionType, prevVer, remoteVer, progressCallback)
	if err != nil {
		return fmt.Errorf("failed to download game patch: %w", err)
	}

	// Verify the patch file exists and is readable
	info, err := os.Stat(pwrPath)
	if err != nil {
		return fmt.Errorf("patch file not accessible: %w", err)
	}
	fmt.Printf("Patch file size: %d bytes\n", info.Size())

	// Apply the patch
	if progressCallback != nil {
		progressCallback("install", 0, "Applying game patch...", "", "", 0, 0)
	}

	if err := patch.ApplyPWR(ctx, versionType, pwrPath, installDirName, progressCallback); err != nil {
		return fmt.Errorf("failed to apply game patch: %w", err)
	}

	// Verify installation
	if _, err := os.Stat(clientPath); err != nil {
		return fmt.Errorf("game installation incomplete: client executable not found at %s", clientPath)
	}

	// Save the new version
	if err := patch.SaveLocalVersion(versionType, remoteVer); err != nil {
		fmt.Printf("Warning: failed to save version info: %v\n", err)
	}

	// Apply online fix only on windows if enabled
	if runtime.GOOS == "windows" && enableOnlineFix {
		if progressCallback != nil {
			progressCallback("online-fix", 0, "Applying online fix...", "", "", 0, 0)
		}

		if err := ApplyOnlineFixWindows(ctx, gameInstallDir, progressCallback); err != nil {
			return fmt.Errorf("failed to apply online fix: %w", err)
		}

		if progressCallback != nil {
			progressCallback("online-fix", 100, "Online fix applied", "", "", 0, 0)
		}
	}

	if progressCallback != nil {
		progressCallback("complete", 100, "Game installed successfully", "", "", 0, 0)
	}

	return nil
}

func getFirstURL(urls []string) string {
	if len(urls) == 0 {
		return "none"
	}
	return urls[0]
}

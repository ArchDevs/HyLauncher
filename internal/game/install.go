package game

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"HyLauncher/internal/env"
	"HyLauncher/internal/java"
	"HyLauncher/internal/pwr"
	"HyLauncher/internal/pwr/butler"
)

func EnsureInstalled(ctx context.Context, progress func(stage string, progress float64, msg string, file string, speed string, down, total int64)) error {
	// Test server connection first
	if progress != nil {
		progress("connection", 5, "Testing server connection...", "", "", 0, 0)
	}

	if err := pwr.TestConnection(); err != nil {
		return fmt.Errorf("cannot connect to game server: %w\n\nPlease check:\n• Your internet connection\n• Firewall/antivirus settings\n• VPN if using one", err)
	}

	// Download JRE
	if err := java.DownloadJRE(ctx, progress); err != nil {
		return fmt.Errorf("failed to download Java Runtime: %w", err)
	}

	// Install Butler
	if _, err := butler.InstallButler(ctx, progress); err != nil {
		return fmt.Errorf("failed to install Butler tool: %w", err)
	}

	// Find latest version with details
	if progress != nil {
		progress("version", 10, "Checking for game updates...", "", "", 0, 0)
	}

	result := pwr.FindLatestVersionWithDetails("release")

	if result.Error != nil {
		// Provide detailed error with platform info
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
				"Version type: release\n\n"+
				"This usually means:\n"+
				"• The game is not yet available for your platform\n"+
				"• Your system architecture is not supported\n"+
				"• Server configuration has changed\n\n"+
				"Please check the official Hytale website for platform availability.",
			runtime.GOOS,
			runtime.GOARCH,
		)
	}

	fmt.Printf("Found latest version: %d\n", result.LatestVersion)
	fmt.Printf("Success URL: %s\n", result.SuccessURL)

	return InstallGame(ctx, "release", result.LatestVersion, progress)
}

func InstallGame(ctx context.Context, versionType string, remoteVer int, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	localStr := pwr.GetLocalVersion()
	local, _ := strconv.Atoi(localStr)

	gameLatestDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "game", "latest")

	// Determine game client executable name
	gameClient := "HytaleClient"
	if runtime.GOOS == "windows" {
		gameClient += ".exe"
	}
	clientPath := filepath.Join(gameLatestDir, "Client", gameClient)
	_, clientErr := os.Stat(clientPath)

	// If we have the right version and the client exists, we're good
	if local == remoteVer && clientErr == nil {
		if progressCallback != nil {
			progressCallback("game", 100, "Game is up to date", "", "", 0, 0)
		}
		return nil
	}

	// Determine if this is a fresh install or update
	prevVer := local
	if clientErr != nil {
		// Client doesn't exist, do full install
		prevVer = 0
		if progressCallback != nil {
			progressCallback("game", 15, fmt.Sprintf("Installing game version %d...", remoteVer), "", "", 0, 0)
		}
	} else {
		if progressCallback != nil {
			progressCallback("game", 15, fmt.Sprintf("Updating from version %d to %d...", local, remoteVer), "", "", 0, 0)
		}
	}

	// Download the patch file
	pwrPath, err := pwr.DownloadPWR(ctx, versionType, prevVer, remoteVer, progressCallback)
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
	if err := pwr.ApplyPWR(ctx, pwrPath, progressCallback); err != nil {
		return fmt.Errorf("failed to apply game patch: %w", err)
	}

	// Verify installation
	if _, err := os.Stat(clientPath); err != nil {
		return fmt.Errorf("game installation incomplete: client executable not found at %s", clientPath)
	}

	// Save the new version
	if err := pwr.SaveLocalVersion(remoteVer); err != nil {
		fmt.Printf("Warning: failed to save version info: %v\n", err)
	}

	if progressCallback != nil {
		progressCallback("game", 100, "Game installed successfully", "", "", 0, 0)
	}

	return nil
}

func getFirstURL(urls []string) string {
	if len(urls) == 0 {
		return "none"
	}
	return urls[0]
}

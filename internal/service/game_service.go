package service

import (
	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/game"
	"HyLauncher/internal/java"
	"HyLauncher/internal/patch"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/fileutil"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	installMutex sync.Mutex
	isInstalling bool
)

type GameService struct {
	ctx      context.Context
	reporter *progress.Reporter
}

func NewGameService(ctx context.Context, reporter *progress.Reporter) *GameService {
	return &GameService{
		ctx:      ctx,
		reporter: reporter,
	}
}

func (s *GameService) VerifyGame(branch string) error {
	s.reporter.Report(progress.StageVerify, 0, "Starting verifying game installation...")

	err := java.VerifyJRE(branch)
	if err != nil {
		return err
	}

	s.reporter.Report(progress.StageVerify, 30, "JRE is installed...")

	err = patch.VerifyButler()
	if err != nil {
		return err
	}

	s.reporter.Report(progress.StageVerify, 65, "Butler is installed...")

	err = game.CheckInstalled(branch)
	if err != nil {
		return err
	}

	s.reporter.Report(progress.StageVerify, 100, "Hytale is installed...")
	s.reporter.Report(progress.StageComplete, 0, "Launcher completed checking, everything is installed")
	return nil
}

func (s *GameService) EnsureInstalled(ctx context.Context, branch string, reporter *progress.Reporter) error {
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

	versionRes := make(chan int, 1)

	if reporter != nil {
		reporter.Report(progress.StageVerify, 0, "Checking for game updates")
	}

	if s.VerifyGame(branch) == nil {
		return nil
	}

	go func() {
		latestVersion, err := patch.FindLatestVersion(branch)
		if err != nil {
			fmt.Println("Can not find latest version:", err)
			latestVersion = 0
		}
		versionRes <- latestVersion
	}()

	var latestVersion int
	select {
	case latestVersion = <-versionRes:
	case <-ctx.Done():
		return ctx.Err()
	}

	if err := java.EnsureJRE(ctx, branch, reporter); err != nil {
		return fmt.Errorf("failed to install jre: %w", err)
	}

	if err := patch.EnsureButler(ctx, reporter); err != nil {
		return fmt.Errorf("failed to install Butler tool: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StageVerify, 100, "Checking complete")
		reporter.Report(progress.StageComplete, 0, fmt.Sprintf("Found version %d", versionRes))
	}

	fmt.Printf("Found latest version: %d\n", latestVersion)

	return s.Install(ctx, branch, latestVersion, reporter)
}

func (s *GameService) Install(ctx context.Context, branch string, latestVersion int, reporter *progress.Reporter) error {
	local, err := config.GetLocalGameVersion()
	if err != nil {
		return fmt.Errorf("Could not fetch local game version")
	}

	gameLatestDir := filepath.Join(env.GetDefaultAppDir(), branch, "package", "game", "latest")

	clientPath, clientErr := fileutil.GetNativeFile(filepath.Join(gameLatestDir, "Client", "HytaleClient"))

	// Check if our game version is same as latest
	if local == latestVersion {
		if reporter != nil {
			reporter.Report(progress.StageComplete, 100, "Game is up to date")
		}
		return nil
	}

	prevVer := local
	if clientErr != nil {
		prevVer = 0
		if reporter != nil {
			reporter.Report(progress.StagePWR, 0, fmt.Sprintf("Installing game version %d...", latestVersion))
		}
	} else {
		if reporter != nil {
			reporter.Report(progress.StagePWR, 0, fmt.Sprintf("Updating from version %d to %d...", local, latestVersion))
		}
	}

	// Download the patch file
	pwrPath, err := patch.DownloadPWR(ctx, branch, prevVer, latestVersion, reporter)
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
	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Applying game patch...")
	}

	if err := patch.ApplyPWR(ctx, pwrPath, branch, reporter); err != nil {
		return fmt.Errorf("failed to apply game patch: %w", err)
	}

	// Verify installation
	if fileutil.FileExists(clientPath) == false {
		return fmt.Errorf("game installation incomplete: client executable not found at %s", clientPath)
	}

	// Save the new version
	if err := config.SaveLocalGameVersion(latestVersion); err != nil {
		fmt.Printf("Warning: failed to save version info: %v\n", err)
	}

	// Apply online fix only on windows
	if runtime.GOOS == "windows" {
		if reporter != nil {
			reporter.Report(progress.StageOnlineFix, 0, "Applying online fix...")
		}

		if err := game.ApplyOnlineFixWindows(ctx, gameLatestDir, reporter); err != nil {
			return fmt.Errorf("failed to apply online fix: %w", err)
		}

		if reporter != nil {
			reporter.Report(progress.StageOnlineFix, 100, "Online fix applied")
		}
	}

	if reporter != nil {
		reporter.Report(progress.StageComplete, 100, "Game installed successfully")
	}

	return nil
}

func (s *GameService) Launch(playerName, branch string) error {
	baseDir := env.GetDefaultAppDir()
	gameDir := filepath.Join(baseDir, branch, "package", "game", "latest")
	userDataDir := filepath.Join(baseDir, "UserData")

	if err := game.EnsureServerAndClientFix(context.Background(), branch, nil); err != nil {
		return err
	}

	clientPath, err := fileutil.GetNativeFile(filepath.Join(gameDir, "Client", "HytaleClient"))
	if err != nil {
		return err
	}

	javaBin, err := java.GetJavaExec(branch)
	if err != nil {
		return err
	}

	_ = os.MkdirAll(userDataDir, 0755)

	playerUUID := game.OfflineUUID(playerName).String()

	cmd := exec.Command(clientPath,
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", javaBin,
		"--auth-mode", "offline",
		"--uuid", playerUUID,
		"--name", playerName,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	game.SetSDLVideoDriver(cmd)

	fmt.Printf(
		"Launching %s (latest) with UUID %s\n",
		playerName,
		playerUUID,
	)

	return cmd.Start()
}

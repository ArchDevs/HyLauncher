package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/game"
	"HyLauncher/internal/java"
	"HyLauncher/internal/patch"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/model"
)

type GameService struct {
	ctx      context.Context
	reporter *progress.Reporter

	installMutex sync.Mutex
}

func NewGameService(ctx context.Context, reporter *progress.Reporter) *GameService {
	return &GameService{ctx: ctx, reporter: reporter}
}

func (s *GameService) VerifyGame(request model.InstanceModel) error {
	s.reporter.Report(progress.StageVerify, 0, "Starting verifying game installation...")

	if err := java.VerifyJRE(request.Branch); err != nil {
		return fmt.Errorf("verify jre: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 30, "JRE is installed...")

	if err := patch.VerifyButler(); err != nil {
		return fmt.Errorf("verify butler: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 65, "Butler is installed...")

	if err := game.CheckInstalled(request.Branch, request.BuildVersion); err != nil {
		return fmt.Errorf("verify game: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 100, "Hytale is installed...")
	return nil
}

func (s *GameService) EnsureInstalled(ctx context.Context, request model.InstanceModel, reporter *progress.Reporter) error {
	s.installMutex.Lock()
	defer s.installMutex.Unlock()

	if reporter != nil {
		reporter.Report(progress.StageVerify, 0, "Checking for game updates")
	}

	if s.VerifyGame(request) == nil {
		return nil
	}

	latestVersion, err := s.fetchLatestVersion(ctx, request.Branch)
	if err != nil {
		return err
	}

	if err := java.EnsureJRE(ctx, request.Branch, reporter); err != nil {
		return fmt.Errorf("install jre: %w", err)
	}

	if err := patch.EnsureButler(ctx, reporter); err != nil {
		return fmt.Errorf("install butler: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StageVerify, 100, "Checking complete")
		reporter.Report(progress.StageComplete, 0, fmt.Sprintf("Found version %d", latestVersion))
	}

	return s.Install(ctx, latestVersion, request, reporter)
}

func (s *GameService) fetchLatestVersion(ctx context.Context, branch string) (int, error) {
	versionChan := make(chan int, 1)
	errChan := make(chan error, 1)

	go func() {
		version, err := patch.FindLatestVersion(branch)
		if err != nil {
			errChan <- err
			return
		}
		versionChan <- version
	}()

	select {
	case version := <-versionChan:
		return version, nil
	case err := <-errChan:
		return 0, fmt.Errorf("find latest version: %w", err)
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

func (s *GameService) Install(ctx context.Context, latestVersion int, request model.InstanceModel, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)
	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)

	pwrPath, err := patch.DownloadPWR(ctx, request.Branch, request.BuildVersion, reporter)
	if err != nil {
		return fmt.Errorf("download patch: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Applying game patch...")
	}

	if err := patch.ApplyPWR(ctx, pwrPath, request, reporter); err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	if runtime.GOOS == "darwin" {
		appPath := filepath.Join(gameDir, "Client", "Hytale.app")
		if !fileutil.FileExists(appPath) {
			return fmt.Errorf("client app not found at %s", appPath)
		}
	} else {
		if !fileutil.FileExists(clientPath) {
			return fmt.Errorf("client executable not found at %s", clientPath)
		}
	}

	config.UpdateInstance("default", func(cfg *config.InstanceConfig) error {
		cfg.Build = request.BuildVersion
		return nil
	})

	if runtime.GOOS == "windows" {
		if reporter != nil {
			reporter.Report(progress.StageOnlineFix, 0, "Applying online fix...")
		}

		if err := game.ApplyOnlineFixWindows(ctx, gameDir, reporter); err != nil {
			return fmt.Errorf("apply online fix: %w", err)
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

func (s *GameService) Launch(playerName string, request model.InstanceModel) error {
	if s.reporter != nil {
		s.reporter.Reset()
		s.reporter.Report(progress.StageLaunch, 0, "Launching game...")
	}

	// Game files are in shared directory
	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)

	// Instance-specific UserData
	userDataDir := env.GetInstanceUserDataDir(request.InstanceID)

	if !fileutil.FileExists(userDataDir) {
		if err := os.MkdirAll(userDataDir, 0755); err != nil {
			return fmt.Errorf("create user data dir: %w", err)
		}
	}

	if err := game.EnsureServerAndClientFix(context.Background(), request, nil); err != nil {
		return fmt.Errorf("apply game fixes: %w", err)
	}

	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)
	javaBin, err := java.GetJavaExec(request.Branch)
	if err != nil {
		return fmt.Errorf("find java: %w", err)
	}

	if runtime.GOOS == "darwin" {
		_ = os.Chmod(clientPath, 0755)
		_ = os.Chmod(javaBin, 0755)
	}

	clientPath, _ = filepath.Abs(clientPath)
	javaBin, _ = filepath.Abs(javaBin)
	userDataDir, _ = filepath.Abs(userDataDir)
	gameDir, _ = filepath.Abs(gameDir)

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

	fmt.Println(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start game process: %w", err)
	}

	return nil
}

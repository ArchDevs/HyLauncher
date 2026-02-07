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

	authSvc *AuthService

	authDomain string
}

func NewGameService(ctx context.Context, reporter *progress.Reporter, svc *AuthService) *GameService {
	return &GameService{
		ctx:        ctx,
		reporter:   reporter,
		authSvc:    svc,
		authDomain: "sanasol.ws",
	}
}

func (s *GameService) SetAuthDomain(domain string) {
	s.authDomain = domain
}

func (s *GameService) EnsureGame(request model.InstanceModel) error {
	s.reporter.Report(progress.StageVerify, 0, "Starting verifying game installation...")

	if err := java.EnsureJRE(s.ctx, request.Branch, s.reporter); err != nil {
		return fmt.Errorf("verify jre: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 30, "JRE is installed...")

	if err := patch.EnsureButler(s.ctx, s.reporter); err != nil {
		return fmt.Errorf("verify butler: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 65, "Butler is installed...")

	if err := game.CheckInstalled(s.ctx, request.Branch, request.BuildVersion); err != nil {
		return fmt.Errorf("verify game: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 100, "Hytale is installed...")
	return nil
}

func (s *GameService) EnsureInstalled(ctx context.Context, request model.InstanceModel, reporter *progress.Reporter) (int, error) {
	s.installMutex.Lock()
	defer s.installMutex.Unlock()

	if reporter != nil {
		reporter.Report(progress.StageVerify, 0, "Checking for game updates")
	}

	if err := s.EnsureGame(request); err == nil {
		return request.BuildVersion, nil
	} else {
		fmt.Println("[EnsureInstalled] verify failed:", err)
		reporter.Report(progress.StageVerify, 0, fmt.Sprintf("Verification failed: %v", err))
	}

	if reporter != nil {
		reporter.Report(progress.StageVerify, 100, "Checking complete")
		reporter.Report(progress.StageComplete, 0, fmt.Sprintf("Found version %d", request.BuildVersion))
	}

	if err := s.Install(ctx, request, reporter); err != nil {
		return 0, err
	}
	return request.BuildVersion, nil
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

func (s *GameService) Install(ctx context.Context, request model.InstanceModel, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)

	pwrPath, sigPath, err := patch.DownloadPWR(ctx, request.Branch, request.BuildVersion, reporter)
	if err != nil {
		return fmt.Errorf("download patch: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Applying game patch...")
	}

	if err := patch.ApplyPWR(ctx, pwrPath, sigPath, request.Branch, request.BuildVersion, reporter); err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)

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

	if err := config.UpdateInstance("default", func(cfg *config.InstanceConfig) error {
		cfg.Build = request.BuildVersion
		return nil
	}); err != nil {
		return fmt.Errorf("update instance: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Applying custom authentication patch...")
	}

	patchRequest := model.InstanceModel{
		InstanceID:   request.InstanceID,
		BuildVersion: request.BuildVersion,
		Branch:       request.Branch,
	}

	if err := patch.EnsureGamePatched(ctx, patchRequest, s.authDomain, reporter); err != nil {
		fmt.Printf("Warning: Failed to apply custom auth patch: %v\n", err)
		fmt.Println("Game will work with official Hytale servers only")
	} else {
		if reporter != nil {
			reporter.Report(progress.StagePatch, 100, "Custom authentication configured")
		}
	}

	if reporter != nil {
		reporter.Report(progress.StageComplete, 100, "Game installed successfully")
	}

	return nil
}

func (s *GameService) Update(ctx context.Context, latestVersion int, request model.InstanceModel, reporter *progress.Reporter) {

}

func (s *GameService) Launch(playerName string, request model.InstanceModel) error {
	gameSession, err := s.authSvc.FetchGameSession(playerName)
	if err != nil {
		return err
	}

	s.reporter.Report(progress.StageLaunch, 0, "Launching game...")

	// Game files are in shared directory
	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)

	// Instance-specific UserData
	userDataDir := env.GetInstanceUserDataDir(request.InstanceID)

	if !fileutil.FileExists(userDataDir) {
		if err := os.MkdirAll(userDataDir, 0755); err != nil {
			return fmt.Errorf("create user data dir: %w", err)
		}
	}

	s.reporter.Report(progress.StageLaunch, 30, "Ensuring game patch...")

	if err := patch.EnsureGamePatched(s.ctx, request, s.authDomain, nil); err != nil {
		fmt.Printf("Warning: Failed to ensure game patch: %v\n", err)
	}

	s.reporter.Report(progress.StageLaunch, 60, "Looking for files...")

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

	s.reporter.Report(progress.StageLaunch, 80, "Launching...")

	cmd := exec.Command(clientPath,
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", javaBin,
		"--auth-mode", "authenticated",
		"--uuid", gameSession.UUID,
		"--name", gameSession.Username,
		"--identity-token", gameSession.IdentityToken,
		"--session-token", gameSession.SessionToken,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	game.SetSDLVideoDriver(cmd)

	fmt.Println(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start game process: %w", err)
	}

	s.reporter.Report(progress.StageLaunch, 100, "Game launched!")
	s.reporter.Reset()

	return nil
}

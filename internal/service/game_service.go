package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"HyLauncher/internal/config"
	"HyLauncher/internal/env"
	"HyLauncher/internal/game"
	"HyLauncher/internal/java"
	"HyLauncher/internal/patch"
	"HyLauncher/internal/platform"
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

	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)
	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)
	fmt.Printf("[EnsureGame] Checking game installation: branch=%s version=%s\n", request.Branch, request.BuildVersion)
	fmt.Printf("[EnsureGame] Game dir: %s\n", gameDir)
	fmt.Printf("[EnsureGame] Client path: %s\n", clientPath)

	if err := game.CheckInstalled(s.ctx, request.Branch, request.BuildVersion); err != nil {
		fmt.Printf("[EnsureGame] CheckInstalled failed: %v\n", err)
		return fmt.Errorf("verify game: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 100, "Hytale is installed...")
	return nil
}

func (s *GameService) EnsureInstalled(ctx context.Context, request model.InstanceModel, reporter *progress.Reporter) (string, error) {
	s.installMutex.Lock()
	defer s.installMutex.Unlock()

	if reporter != nil {
		reporter.Report(progress.StageVerify, 0, "Checking for game updates")
	}

	if err := s.EnsureGame(request); err == nil {
		return request.BuildVersion, nil
	} else {
		fmt.Println("[EnsureInstalled] verify failed:", err)
		if reporter != nil {
			reporter.Report(progress.StageVerify, 0, fmt.Sprintf("Verification failed: %v", err))
		}
	}

	if request.BuildVersion == "auto" {
		latest, err := s.fetchLatestVersion(ctx, request.Branch)
		if err != nil {
			return "", fmt.Errorf("fetch latest version: %w", err)
		}

		autoDir := env.GetGameDir(request.Branch, "auto")
		versionFile := filepath.Join(autoDir, ".version")
		currentVer := 0
		if data, err := os.ReadFile(versionFile); err == nil {
			currentVer, _ = strconv.Atoi(string(data))
		}

		fmt.Printf("[EnsureInstalled] Auto version check: current=%d latest=%d versionFile=%s\n", currentVer, latest, versionFile)
		
		checkErr := game.CheckInstalled(ctx, request.Branch, "auto")
		fmt.Printf("[EnsureInstalled] CheckInstalled result: %v\n", checkErr)

		if currentVer == latest && checkErr == nil {
			if reporter != nil {
				reporter.Report(progress.StageVerify, 100, "Auto build is up to date")
			}
			return "auto", nil
		}
		
		fmt.Printf("[EnsureInstalled] Reinstalling: currentVer=%d latest=%d checkErr=%v\n", currentVer, latest, checkErr)

		if reporter != nil {
			reporter.Report(progress.StageVerify, 50, fmt.Sprintf("Updating auto build to version %d", latest))
		}

		if err := s.installInternal(ctx, request.Branch, "auto", latest, reporter); err != nil {
			return "", err
		}

		_ = os.WriteFile(versionFile, []byte(strconv.Itoa(latest)), 0644)

		return "auto", nil
	}

	verInt, err := strconv.Atoi(request.BuildVersion)
	if err != nil {
		return "", fmt.Errorf("invalid static version %q: %w", request.BuildVersion, err)
	}

	if err := s.installInternal(ctx, request.Branch, request.BuildVersion, verInt, reporter); err != nil {
		return "", err
	}

	if err := config.UpdateInstance(request.InstanceID, func(cfg *config.InstanceConfig) error {
		cfg.Build = request.BuildVersion
		return nil
	}); err != nil {
		return "", fmt.Errorf("update instance: %w", err)
	}

	return request.BuildVersion, nil
}

func (s *GameService) installInternal(ctx context.Context, branch string, version string, verInt int, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(branch, version)

	pwrPath, sigPath, err := patch.DownloadPWR(ctx, branch, verInt, reporter)
	if err != nil {
		return fmt.Errorf("download patch: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Applying game patch...")
	}

	if err := patch.ApplyPWR(ctx, pwrPath, sigPath, branch, version, reporter); err != nil {
		return fmt.Errorf("apply patch: %w", err)
	}

	// Fix permissions on macOS after patching
	if runtime.GOOS == "darwin" {
		clientExec := filepath.Join(gameDir, "Client", "Hytale.app", "Contents", "MacOS", "HytaleClient")
		if fileutil.FileExists(clientExec) {
			_ = os.Chmod(clientExec, 0755)
		}
		// Also fix the Java runtime if needed
		jreDir := env.GetJREDir()
		javaExec := filepath.Join(jreDir, "bin", "java")
		if fileutil.FileExists(javaExec) {
			_ = os.Chmod(javaExec, 0755)
		}
	}

	clientPath := env.GetGameClientPath(branch, version)

	if runtime.GOOS == "darwin" {
		// On macOS, check for the executable inside the app bundle
		if clientPath == "" || !fileutil.FileExists(clientPath) {
			return fmt.Errorf("client executable not found in app bundle")
		}
	} else {
		if !fileutil.FileExists(clientPath) {
			return fmt.Errorf("client executable not found at %s", clientPath)
		}
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Applying custom authentication patch...")
	}

	patchRequest := model.InstanceModel{
		BuildVersion: version,
		Branch:       branch,
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

func (s *GameService) Install(ctx context.Context, request model.InstanceModel, reporter *progress.Reporter) error {
	verInt, err := strconv.Atoi(request.BuildVersion)
	if err != nil {
		return fmt.Errorf("invalid version for install: %w", err)
	}

	if err := s.installInternal(ctx, request.Branch, request.BuildVersion, verInt, reporter); err != nil {
		return err
	}

	return nil
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

	fmt.Printf("[Launch] Starting game: %s\n", clientPath)
	fmt.Printf("[Launch] Working dir: %s\n", gameDir)
	fmt.Printf("[Launch] Command: %v\n", cmd.Args)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start game process: %w", err)
	}

	fmt.Printf("[Launch] Process started with PID: %d\n", cmd.Process.Pid)

	// On macOS, detach the process so it doesn't die when launcher exits
	if runtime.GOOS == "darwin" && cmd.Process != nil {
		if err := cmd.Process.Release(); err != nil {
			fmt.Printf("[Launch] Warning: could not detach process: %v\n", err)
		} else {
			fmt.Printf("[Launch] Process detached\n")
		}
	}

	// On macOS, remove quarantine from the binary to prevent silent killing
	if runtime.GOOS == "darwin" {
		if err := platform.RemoveQuarantine(clientPath); err != nil {
			fmt.Printf("[Launch] Warning: could not remove quarantine: %v\n", err)
		}
	}

	// Wait a moment and check if process is still running
	time.Sleep(500 * time.Millisecond)
	if cmd.Process != nil {
		// Signal 0 is a no-op that checks if process exists (Unix only)
		if runtime.GOOS != "windows" {
			if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
				fmt.Printf("[Launch] Process already exited: %v\n", err)
				// Try to get exit code
				if waitErr := cmd.Wait(); waitErr != nil {
					fmt.Printf("[Launch] Process exit error: %v\n", waitErr)
					return fmt.Errorf("game process exited immediately: %w", waitErr)
				}
				return fmt.Errorf("game process exited immediately")
			}
		}
		fmt.Printf("[Launch] Process is still running (PID: %d)\n", cmd.Process.Pid)
	}

	s.reporter.Report(progress.StageLaunch, 100, "Game launched!")
	s.reporter.Reset()

	return nil
}

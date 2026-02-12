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

	latest, err := s.fetchLatestVersion(ctx, request.Branch)
	if err != nil {
		return "", fmt.Errorf("fetch latest version: %w", err)
	}

	if request.BuildVersion == "auto" {
		// AUTO mode: always stay on latest with incremental updates
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

	if request.BuildVersion == "latest" {
		latestDir := env.GetGameDir(request.Branch, "latest")
		versionFile := filepath.Join(latestDir, ".version")

		// Check if already installed with correct version
		if data, err := os.ReadFile(versionFile); err == nil {
			if installedVer, _ := strconv.Atoi(string(data)); installedVer == latest {
				if checkErr := game.CheckInstalled(ctx, request.Branch, "latest"); checkErr == nil {
					if reporter != nil {
						reporter.Report(progress.StageVerify, 100, "Latest build is up to date")
					}
					return "latest", nil
				}
			}
		}

		fmt.Printf("[EnsureInstalled] Installing latest version: %d\n", latest)

		if reporter != nil {
			reporter.Report(progress.StageVerify, 50, fmt.Sprintf("Installing latest version %d", latest))
		}

		if err := s.installInternal(ctx, request.Branch, "latest", latest, reporter); err != nil {
			return "", err
		}

		_ = os.WriteFile(versionFile, []byte(strconv.Itoa(latest)), 0644)

		return "latest", nil
	}

	return "", fmt.Errorf("invalid version %q: only 'auto' and 'latest' are supported", request.BuildVersion)
}

func (s *GameService) installInternal(ctx context.Context, branch string, version string, verInt int, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(branch, version)

	// Determine current version for incremental patching
	currentVer := 0
	if version == "auto" {
		// For AUTO version, read current version from .version file
		autoDir := env.GetGameDir(branch, "auto")
		versionFile := filepath.Join(autoDir, ".version")
		if data, err := os.ReadFile(versionFile); err == nil {
			currentVer, _ = strconv.Atoi(string(data))
		}
	}
	// For specific versions, currentVer remains 0 (fresh install from version 0)

	if err := patch.DownloadAndApplyPWR(ctx, branch, currentVer, verInt, reporter); err != nil {
		return fmt.Errorf("download and apply patches: %w", err)
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

func (s *GameService) Launch(playerName string, request model.InstanceModel, serverIP ...string) error {
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

	args := []string{
		"--app-dir", gameDir,
		"--user-dir", userDataDir,
		"--java-exec", javaBin,
		"--auth-mode", "authenticated",
		"--uuid", gameSession.UUID,
		"--name", gameSession.Username,
		"--identity-token", gameSession.IdentityToken,
		"--session-token", gameSession.SessionToken,
	}

	if len(serverIP) > 0 && serverIP[0] != "" {
		args = append(args, "--server", serverIP[0])
	}

	cmd := exec.Command(clientPath, args...)

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

	if runtime.GOOS == "darwin" && cmd.Process != nil {
		if err := cmd.Process.Release(); err != nil {
			fmt.Printf("[Launch] Warning: could not detach process: %v\n", err)
		} else {
			fmt.Printf("[Launch] Process detached\n")
		}
	}

	if runtime.GOOS == "darwin" {
		if err := platform.RemoveQuarantine(clientPath); err != nil {
			fmt.Printf("[Launch] Warning: could not remove quarantine: %v\n", err)
		}
	}

	time.Sleep(500 * time.Millisecond)
	if cmd.Process != nil {
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

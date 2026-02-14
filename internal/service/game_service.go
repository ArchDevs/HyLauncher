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
	"HyLauncher/pkg/logger"
	"HyLauncher/pkg/model"
)

type GameService struct {
	ctx        context.Context
	reporter   *progress.Reporter
	authSvc    *AuthService
	authDomain string
	installMu  sync.Mutex
}

func NewGameService(ctx context.Context, reporter *progress.Reporter, svc *AuthService) *GameService {
	return &GameService{
		ctx:        ctx,
		reporter:   reporter,
		authSvc:    svc,
		authDomain: "sanasol.ws",
	}
}

func (s *GameService) EnsureGame(request model.InstanceModel) error {
	s.reporter.Report(progress.StageVerify, 0, "Verifying installation...")

	if err := java.EnsureJRE(s.ctx, request.Branch, s.reporter); err != nil {
		return fmt.Errorf("jre: %w", err)
	}

	if err := patch.EnsureButler(s.ctx, s.reporter); err != nil {
		return fmt.Errorf("butler: %w", err)
	}

	if err := game.CheckInstalled(s.ctx, request.Branch, request.BuildVersion); err != nil {
		return fmt.Errorf("game files: %w", err)
	}

	s.reporter.Report(progress.StageVerify, 100, "Ready")
	return nil
}

func (s *GameService) EnsureInstalled(ctx context.Context, request model.InstanceModel, reporter *progress.Reporter) (string, error) {
	s.installMu.Lock()
	defer s.installMu.Unlock()

	if reporter != nil {
		reporter.Report(progress.StageVerify, 0, "Checking for updates...")
	}

	latest, err := patch.FindLatestVersion(request.Branch)
	if err != nil {
		return "", fmt.Errorf("fetch latest: %w", err)
	}

	switch request.BuildVersion {
	case "auto":
		return s.handleAutoVersion(ctx, request.Branch, latest, reporter)
	case "latest":
		return s.handleLatestVersion(ctx, request.Branch, latest, reporter)
	default:
		if err := s.EnsureGame(request); err == nil {
			return request.BuildVersion, nil
		}
		return "", fmt.Errorf("version %q not installed", request.BuildVersion)
	}
}

func (s *GameService) handleAutoVersion(ctx context.Context, branch string, latest int, reporter *progress.Reporter) (string, error) {
	autoDir := env.GetGameDir(branch, "auto")
	versionFile := filepath.Join(autoDir, ".version")

	currentVer := s.readVersionFile(versionFile)

	if currentVer == latest && game.CheckInstalled(ctx, branch, "auto") == nil {
		if reporter != nil {
			reporter.Report(progress.StageVerify, 100, "Up to date")
		}
		return "auto", nil
	}

	if reporter != nil {
		reporter.Report(progress.StageVerify, 50, fmt.Sprintf("Updating to %d...", latest))
	}

	if err := s.install(ctx, branch, "auto", latest, reporter); err != nil {
		return "", err
	}

	_ = os.WriteFile(versionFile, []byte(strconv.Itoa(latest)), 0644)
	return "auto", nil
}

func (s *GameService) handleLatestVersion(ctx context.Context, branch string, latest int, reporter *progress.Reporter) (string, error) {
	versionStr := strconv.Itoa(latest)

	if game.CheckInstalled(ctx, branch, versionStr) == nil {
		if reporter != nil {
			reporter.Report(progress.StageVerify, 100, "Up to date")
		}
		return versionStr, nil
	}

	if reporter != nil {
		reporter.Report(progress.StageVerify, 50, fmt.Sprintf("Installing %d...", latest))
	}

	if err := s.install(ctx, branch, versionStr, latest, reporter); err != nil {
		return "", err
	}

	return versionStr, nil
}

func (s *GameService) readVersionFile(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	ver, _ := strconv.Atoi(string(data))
	return ver
}

func (s *GameService) install(ctx context.Context, branch, version string, targetVer int, reporter *progress.Reporter) error {
	currentVer := 0
	if version == "auto" {
		versionFile := filepath.Join(env.GetGameDir(branch, "auto"), ".version")
		currentVer = s.readVersionFile(versionFile)
	}

	if err := patch.DownloadAndApplyPWR(ctx, branch, currentVer, targetVer, version, reporter); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	if err := s.fixPermissions(branch, version); err != nil {
		return err
	}

	if err := s.applyAuthPatch(branch, version, reporter); err != nil {
		logger.Warn("Auth patch failed", "error", err)
	}

	if reporter != nil {
		reporter.Report(progress.StageComplete, 100, "Done")
	}
	return nil
}

func (s *GameService) fixPermissions(branch, version string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}

	gameDir := env.GetGameDir(branch, version)
	clientExec := filepath.Join(gameDir, "Client", "Hytale.app", "Contents", "MacOS", "HytaleClient")

	if fileutil.FileExists(clientExec) {
		_ = os.Chmod(clientExec, 0755)
	}

	javaExec := filepath.Join(env.GetJREDir(), "bin", "java")
	if fileutil.FileExists(javaExec) {
		_ = os.Chmod(javaExec, 0755)
	}

	return nil
}

func (s *GameService) applyAuthPatch(branch, version string, reporter *progress.Reporter) error {
	if reporter != nil {
		reporter.Report(progress.StagePatch, 0, "Patching auth...")
	}

	req := model.InstanceModel{BuildVersion: version, Branch: branch}
	if err := patch.EnsureGamePatched(s.ctx, req, s.authDomain, reporter); err != nil {
		return err
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 100, "Auth ready")
	}
	return nil
}

func (s *GameService) Launch(playerName string, request model.InstanceModel, serverIP ...string) error {
	session, err := s.authSvc.FetchGameSession(playerName)
	if err != nil {
		return err
	}

	s.reporter.Report(progress.StageLaunch, 0, "Launching...")

	gameDir := env.GetGameDir(request.Branch, request.BuildVersion)
	userDataDir := env.GetInstanceUserDataDir(request.InstanceID)

	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return fmt.Errorf("userdata: %w", err)
	}

	_ = patch.EnsureGamePatched(s.ctx, request, s.authDomain, nil)

	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)
	if clientPath == "" {
		return fmt.Errorf("client not found")
	}

	javaBin, err := java.GetJavaExec(request.Branch)
	if err != nil {
		return fmt.Errorf("java: %w", err)
	}

	if runtime.GOOS == "darwin" {
		_ = os.Chmod(clientPath, 0755)
		_ = os.Chmod(javaBin, 0755)
	}

	args := []string{
		"--app-dir", mustAbs(gameDir),
		"--user-dir", mustAbs(userDataDir),
		"--java-exec", mustAbs(javaBin),
		"--auth-mode", "authenticated",
		"--uuid", session.UUID,
		"--name", session.Username,
		"--identity-token", session.IdentityToken,
		"--session-token", session.SessionToken,
	}

	if len(serverIP) > 0 && serverIP[0] != "" {
		args = append(args, "--server", serverIP[0])
	}

	cmd := exec.Command(clientPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	game.SetSDLVideoDriver(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}

	if runtime.GOOS == "darwin" {
		_ = cmd.Process.Release()
		_ = platform.RemoveQuarantine(clientPath)
	}

	time.Sleep(500 * time.Millisecond)

	if cmd.Process != nil && runtime.GOOS != "windows" {
		if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
			_ = cmd.Wait()
			return fmt.Errorf("process exited")
		}
	}

	s.reporter.Report(progress.StageLaunch, 100, "Launched!")
	s.reporter.Reset()
	return nil
}

func mustAbs(path string) string {
	abs, _ := filepath.Abs(path)
	return abs
}

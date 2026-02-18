package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/patch"
	"HyLauncher/internal/platform"
	"HyLauncher/internal/service"
	"HyLauncher/pkg/hyerrors"
	"HyLauncher/pkg/logger"
)

func (a *App) GetLogs() (string, error) {
	if a.crashSvc == nil {
		return "", hyerrors.Internal("diagnostics not initialized")
	}
	return a.crashSvc.GetLogs()
}

func (a *App) GetCrashReports() ([]service.CrashReport, error) {
	if a.crashSvc == nil {
		return nil, hyerrors.Internal("diagnostics not initialized")
	}
	return a.crashSvc.GetCrashReports()
}

func (a *App) validatePlayerName(name string) error {
	re := regexp.MustCompile("^[A-Za-z0-9_]{3,16}$")

	if !re.MatchString(name) {
		return hyerrors.Validation("nickname should be 3-16 characters long, consisting only of letters, numbers, and underscores").
			WithContext("length", len(name)).
			WithContext("name", name)
	}

	return nil
}

type ButlerDiagnosticResult struct {
	Success      bool   `json:"success"`
	Version      string `json:"version,omitempty"`
	Error        string `json:"error,omitempty"`
	Details      string `json:"details,omitempty"`
	DebugLogPath string `json:"debugLogPath,omitempty"`
}

func (a *App) TestButler() ButlerDiagnosticResult {
	logger.Info("Starting Butler diagnostic test")

	butlerPath, err := patch.GetButlerExec()
	if err != nil {
		logger.Error("Butler diagnostic: cannot find butler", "error", err)
		return ButlerDiagnosticResult{
			Success: false,
			Error:   fmt.Sprintf("Cannot find butler: %v", err),
		}
	}

	// Test 1: Check butler version
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	versionCmd := exec.CommandContext(ctx, butlerPath, "--version")
	platform.HideConsoleWindow(versionCmd)
	versionOutput, versionErr := versionCmd.CombinedOutput()

	versionStr := string(versionOutput)
	if versionErr != nil {
		logger.Error("Butler diagnostic: version check failed", "error", versionErr, "output", versionStr)
		return ButlerDiagnosticResult{
			Success: false,
			Error:   fmt.Sprintf("Version check failed: %v", versionErr),
			Details: versionStr,
		}
	}

	logger.Info("Butler diagnostic: version check passed", "version", versionStr)

	// Test 2: Check butler can access filesystem
	testDir := filepath.Join(env.GetCacheDir(), "butler-test")
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Create a simple test file
	testFile := filepath.Join(testDir, "test.txt")
	testContent := []byte("hello world")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		logger.Error("Butler diagnostic: failed to create test file", "error", err)
		return ButlerDiagnosticResult{
			Success: false,
			Version: versionStr,
			Error:   fmt.Sprintf("Failed to create test file: %v", err),
		}
	}

	// Test 3: Check environment
	envInfo := fmt.Sprintf(
		"OS: %s\nArch: %s\nGo Version: %s\nButler Path: %s\nTest Dir: %s\nPATH: %s\nTEMP: %s\nTMPDIR: %s",
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		butlerPath,
		testDir,
		os.Getenv("PATH"),
		os.Getenv("TEMP"),
		os.Getenv("TMPDIR"),
	)

	// Write comprehensive debug log
	debugLogPath := filepath.Join(env.GetCacheDir(), "butler-diagnostic.log")
	debugContent := fmt.Sprintf(
		"=== BUTLER DIAGNOSTIC REPORT ===\n"+
			"Time: %s\n\n"+
			"=== ENVIRONMENT ===\n%s\n\n"+
			"=== VERSION OUTPUT ===\n%s\n",
		time.Now().Format(time.RFC3339),
		envInfo,
		versionStr,
	)

	if err := os.WriteFile(debugLogPath, []byte(debugContent), 0644); err != nil {
		logger.Warn("Butler diagnostic: failed to write debug log", "error", err)
	}

	logger.Info("Butler diagnostic completed successfully", "version", versionStr, "debugLog", debugLogPath)

	return ButlerDiagnosticResult{
		Success:      true,
		Version:      versionStr,
		Details:      envInfo,
		DebugLogPath: debugLogPath,
	}
}

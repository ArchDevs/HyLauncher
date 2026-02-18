package patch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/platform"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/logger"
)

type PatchRequest struct {
	OS      string `json:"os"`
	Arch    string `json:"arch"`
	Branch  string `json:"branch"`
	Version string `json:"version"`
}

type PatchStep struct {
	From    int    `json:"from"`
	To      int    `json:"to"`
	PWR     string `json:"pwr"`
	PWRHead string `json:"pwrHead"`
	Sig     string `json:"sig"`
}

type PatchStepsResponse struct {
	Steps []PatchStep `json:"steps"`
}

func DownloadAndApplyPWR(ctx context.Context, branch string, currentVer int, targetVer int, versionDir string, reporter *progress.Reporter) error {
	logger.Info("Starting patch download", "branch", branch, "from", currentVer, "to", targetVer)

	var pwrPath string

	steps, err := fetchPatchSteps(ctx, branch, currentVer)
	if err != nil {
		logger.Error("Failed to fetch patch steps", "branch", branch, "error", err)
		return fmt.Errorf("fetch patch steps: %w", err)
	}

	if len(steps) == 0 {
		logger.Warn("No patch steps available", "branch", branch, "currentVer", currentVer)
		return fmt.Errorf("no patch steps available")
	}

	logger.Info("Found patch steps", "count", len(steps), "branch", branch)
	for i, step := range steps {
		logger.Info("  Step", "index", i, "from", step.From, "to", step.To)
	}

	for i, step := range steps {
		if targetVer > 0 && step.From >= targetVer {
			logger.Info("Reached target version, stopping", "target", targetVer, "current", step.From)
			break
		}

		logger.Info("Downloading patch", "from", step.From, "to", step.To, "progress", fmt.Sprintf("%d/%d", i+1, len(steps)))

		if reporter != nil {
			reporter.Report(progress.StagePatch, 0, fmt.Sprintf("Patching %d → %d (%d/%d)", step.From, step.To, i+1, len(steps)))
		}

		pwrPath, sigPath, err := downloadPatchStep(ctx, step, reporter)
		if err != nil {
			logger.Error("Failed to download patch", "from", step.From, "to", step.To, "error", err)
			return fmt.Errorf("download patch step %d→%d: %w", step.From, step.To, err)
		}

		logger.Info("Applying patch", "from", step.From, "to", step.To)
		if err := applyPWR(ctx, pwrPath, sigPath, branch, versionDir, reporter); err != nil {
			_ = os.Remove(pwrPath)
			_ = os.Remove(sigPath)
			logger.Error("Failed to apply patch", "from", step.From, "to", step.To, "error", err)
			return fmt.Errorf("apply patch %d→%d: %w", step.From, step.To, err)
		}

		logger.Info("Patch applied successfully", "from", step.From, "to", step.To)
	}

	_ = os.RemoveAll(pwrPath)
	logger.Info("All patches applied", "totalSteps", len(steps))

	return nil
}

func applyPWR(ctx context.Context, pwrFile string, sigFile string, branch string, version string, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(branch, version)
	stagingDir := filepath.Join(env.GetCacheDir(), "staging-temp")
	_ = os.RemoveAll(stagingDir)

	_ = os.MkdirAll(gameDir, 0755)
	_ = os.MkdirAll(stagingDir, 0755)

	// Log what files are currently in the game directory
	logger.Info("Game directory contents before patch", "dir", gameDir)
	entries, _ := os.ReadDir(gameDir)
	for _, entry := range entries {
		logger.Info("  File", "name", entry.Name(), "isDir", entry.IsDir())
	}

	butlerPath, err := GetButlerExec()
	if err != nil {
		return fmt.Errorf("cannot get butler: %w", err)
	}

	versionCmd := exec.CommandContext(ctx, butlerPath, "--version")
	platform.HideConsoleWindow(versionCmd)
	versionOutput, versionErr := versionCmd.CombinedOutput()
	logger.Info("Butler version check", "output", string(versionOutput), "error", versionErr)

	logger.Info("Running butler apply", "pwr", pwrFile, "sig", sigFile, "gameDir", gameDir, "stagingDir", stagingDir)

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		"--signature", sigFile,
		"--verbose",
		pwrFile,
		gameDir,
	)

	platform.HideConsoleWindow(cmd)

	logger.Info("Butler environment",
		"pwd", gameDir,
		"path", os.Getenv("PATH"),
		"temp", os.Getenv("TEMP"),
		"tmpdir", os.Getenv("TMPDIR"))

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if reporter != nil {
		reporter.Report(progress.StagePatch, 60, "Applying game patch...")
	}

	logger.Info("Starting butler apply command", "args", cmd.Args)

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(stagingDir)
		stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()

		debugLogPath := filepath.Join(env.GetCacheDir(), "butler-debug.log")
		debugContent := fmt.Sprintf(
			"=== BUTLER DEBUG LOG ===\n"+
				"Time: %s\n"+
				"Command: %v\n"+
				"Working Dir: %s\n"+
				"Game Dir: %s\n"+
				"Staging Dir: %s\n"+
				"PWR File: %s\n"+
				"SIG File: %s\n"+
				"\n=== ENVIRONMENT ===\n"+
				"PATH=%s\n"+
				"TEMP=%s\n"+
				"TMPDIR=%s\n"+
				"\n=== STDOUT ===\n%s\n"+
				"\n=== STDERR ===\n%s\n"+
				"\n=== ERROR ===\n%v\n",
			time.Now().Format(time.RFC3339),
			cmd.Args,
			gameDir,
			gameDir,
			stagingDir,
			pwrFile,
			sigFile,
			os.Getenv("PATH"),
			os.Getenv("TEMP"),
			os.Getenv("TMPDIR"),
			stdoutStr,
			stderrStr,
			err,
		)
		_ = os.WriteFile(debugLogPath, []byte(debugContent), 0644)
		logger.Error("Butler apply failed", "error", err, "stdout", stdoutStr, "stderr", stderrStr, "debugLog", debugLogPath)

		// Check if it's a signature verification error (user modified files)
		// The error can appear in stdout or stderr and has various formats
		combinedOutput := stdoutStr + stderrStr
		isSignatureError := strings.Contains(combinedOutput, "Verifying against signature") &&
			(strings.Contains(combinedOutput, "expected") || strings.Contains(combinedOutput, "dirs"))

		if isSignatureError {
			logger.Warn("Signature verification failed - game files were modified, will clean and retry",
				"stdout", stdoutStr,
				"stderr", stderrStr)

			// Clean the game directory and retry once
			logger.Info("Cleaning game directory for fresh install", "dir", gameDir)
			if cleanErr := os.RemoveAll(gameDir); cleanErr != nil {
				logger.Error("Failed to clean game directory", "error", cleanErr)
				return fmt.Errorf("signature verification failed and cleanup failed: %w (original error: %v)", cleanErr, err)
			}

			// Recreate the directory
			if mkdirErr := os.MkdirAll(gameDir, 0755); mkdirErr != nil {
				logger.Error("Failed to recreate game directory", "error", mkdirErr)
				return fmt.Errorf("signature verification failed and directory recreation failed: %w (original error: %v)", mkdirErr, err)
			}

			logger.Info("Game directory cleaned, retrying patch application")

			// Retry the patch application
			retryCmd := exec.CommandContext(ctx, butlerPath,
				"apply",
				"--staging-dir", stagingDir,
				"--signature", sigFile,
				"--verbose",
				pwrFile,
				gameDir,
			)
			platform.HideConsoleWindow(retryCmd)

			var retryStdoutBuf, retryStderrBuf bytes.Buffer
			retryCmd.Stdout = &retryStdoutBuf
			retryCmd.Stderr = &retryStderrBuf

			if reporter != nil {
				reporter.Report(progress.StagePatch, 60, "Retrying after cleaning modified files...")
			}

			if retryErr := retryCmd.Run(); retryErr != nil {
				_ = os.RemoveAll(stagingDir)
				logger.Error("Butler apply retry failed after cleanup", "error", retryErr, "stdout", retryStdoutBuf.String(), "stderr", retryStderrBuf.String())
				return fmt.Errorf("patch failed even after cleaning modified files: %w (original error: %v)", retryErr, err)
			}

			logger.Info("Patch applied successfully after cleaning modified files")
			stdoutStr = retryStdoutBuf.String()
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("butler apply failed with exit code %d: stderr=%s debugLog=%s", exitErr.ExitCode(), stderrStr, debugLogPath)
		}
		return fmt.Errorf("butler apply failed: %w (debug log: %s)", err, debugLogPath)
	}

	stdoutStr := stdoutBuf.String()
	logger.Info("Butler apply completed", "stdout", stdoutStr)

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
	}

	_ = os.RemoveAll(stagingDir)

	if reporter != nil {
		reporter.Report(progress.StagePatch, 80, "Game patched!")
	}
	return nil
}

func fetchPatchSteps(ctx context.Context, branch string, currentVer int) ([]PatchStep, error) {
	reqBody := PatchRequest{
		OS:      env.GetOS(),
		Arch:    env.GetArchForAPI(),
		Branch:  branch,
		Version: strconv.Itoa(currentVer),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.hylauncher.fun/v1/pwr", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result PatchStepsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Steps, nil
}

func downloadPatchStep(ctx context.Context, step PatchStep, reporter *progress.Reporter) (pwrPath string, sigPath string, err error) {
	cacheDir := env.GetCacheDir()
	_ = os.MkdirAll(cacheDir, 0755)

	pwrFileName := fmt.Sprintf("%d_to_%d.pwr", step.From, step.To)
	sigFileName := fmt.Sprintf("%d_to_%d.pwr.sig", step.From, step.To)

	pwrDest := filepath.Join(cacheDir, pwrFileName)
	sigDest := filepath.Join(cacheDir, sigFileName)

	_, pwrErr := os.Stat(pwrDest)
	_, sigErr := os.Stat(sigDest)
	if pwrErr == nil && sigErr == nil {
		if reporter != nil {
			reporter.Report(progress.StagePWR, 100, "Patch files cached")
		}
		return pwrDest, sigDest, nil
	}

	if reporter != nil {
		reporter.Report(progress.StagePWR, 0, fmt.Sprintf("Downloading patch %d→%d...", step.From, step.To))
	}

	pwrScaler := progress.NewScaler(reporter, progress.StagePWR, 0, 70)
	if err := download.DownloadWithReporter(ctx, pwrDest, step.PWR, pwrFileName, reporter, progress.StagePWR, pwrScaler); err != nil {
		_ = os.Remove(pwrDest + ".tmp")
		return "", "", fmt.Errorf("download PWR: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePWR, 70, "Downloading signature...")
	}

	sigScaler := progress.NewScaler(reporter, progress.StagePWR, 70, 100)
	if err := download.DownloadWithReporter(ctx, sigDest, step.Sig, sigFileName, reporter, progress.StagePWR, sigScaler); err != nil {
		_ = os.Remove(sigDest + ".tmp")
		return "", "", fmt.Errorf("download signature: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePWR, 100, "Patch files downloaded")
	}

	return pwrDest, sigDest, nil
}

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

	steps, err := fetchPatchSteps(ctx, branch, currentVer)
	if err != nil {
		return fmt.Errorf("fetch patch steps: %w", err)
	}

	if len(steps) == 0 {
		return fmt.Errorf("no patch steps available")
	}

	for i, step := range steps {
		if targetVer > 0 && step.From >= targetVer {
			break
		}

		logger.Info("Downloading patch", "from", step.From, "to", step.To, "progress", fmt.Sprintf("%d/%d", i+1, len(steps)))
		if reporter != nil {
			reporter.Report(progress.StagePatch, 0, fmt.Sprintf("Patching %d → %d (%d/%d)", step.From, step.To, i+1, len(steps)))
		}

		pwrPath, sigPath, err := downloadPatchStep(ctx, step, reporter)
		if err != nil {
			return fmt.Errorf("download patch step %d→%d: %w", step.From, step.To, err)
		}

		if err := applyPWR(ctx, pwrPath, sigPath, branch, versionDir, reporter); err != nil {
			_ = os.Remove(pwrPath)
			_ = os.Remove(sigPath)
			return fmt.Errorf("apply patch %d→%d: %w", step.From, step.To, err)
		}
	}

	logger.Info("All patches applied", "totalSteps", len(steps))
	return nil
}

func applyPWR(ctx context.Context, pwrFile, sigFile, branch, version string, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(branch, version)
	stagingDir := filepath.Join(env.GetCacheDir(), "staging-temp")

	_ = os.RemoveAll(stagingDir)
	_ = os.MkdirAll(gameDir, 0755)
	_ = os.MkdirAll(stagingDir, 0755)

	butlerPath, err := GetButlerExec()
	if err != nil {
		return fmt.Errorf("cannot get butler: %w", err)
	}

	cmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		"--signature", sigFile,
		pwrFile,
		gameDir,
	)
	platform.HideConsoleWindow(cmd)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if reporter != nil {
		reporter.Report(progress.StagePatch, 60, "Applying game patch...")
	}

	if err := cmd.Run(); err != nil {
		return handleApplyError(ctx, err, cmd, stdoutBuf.String(), stderrBuf.String(), gameDir, stagingDir, butlerPath, sigFile, pwrFile, reporter)
	}

	cleanup(cmd, stagingDir, reporter)
	return nil
}

func handleApplyError(ctx context.Context, err error, cmd *exec.Cmd, stdoutStr, stderrStr, gameDir, stagingDir, butlerPath, sigFile, pwrFile string, reporter *progress.Reporter) error {
	_ = os.RemoveAll(stagingDir)

	debugLogPath := writeDebugLog(cmd.Args, gameDir, stdoutStr, stderrStr, err)

	if isSignatureError(stdoutStr, stderrStr) {
		return retryAfterCleanup(ctx, gameDir, stagingDir, butlerPath, sigFile, pwrFile, reporter, err)
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("butler apply failed (exit %d): %s (log: %s)", exitErr.ExitCode(), stderrStr, debugLogPath)
	}
	return fmt.Errorf("butler apply failed: %w (log: %s)", err, debugLogPath)
}

func isSignatureError(stdout, stderr string) bool {
	combined := stdout + stderr
	return strings.Contains(combined, "Verifying against signature") &&
		(strings.Contains(combined, "expected") || strings.Contains(combined, "dirs"))
}

func retryAfterCleanup(ctx context.Context, gameDir, stagingDir, butlerPath, sigFile, pwrFile string, reporter *progress.Reporter, originalErr error) error {
	logger.Warn("Signature verification failed - cleaning and retrying", "gameDir", gameDir)

	_ = os.RemoveAll(gameDir)
	_ = os.RemoveAll(stagingDir)
	_ = os.MkdirAll(gameDir, 0755)
	_ = os.MkdirAll(stagingDir, 0755)

	if reporter != nil {
		reporter.Report(progress.StagePatch, 60, "Retrying after cleaning modified files...")
	}

	retryCmd := exec.CommandContext(ctx, butlerPath,
		"apply",
		"--staging-dir", stagingDir,
		"--signature", sigFile,
		pwrFile,
		gameDir,
	)
	platform.HideConsoleWindow(retryCmd)

	var stdoutBuf, stderrBuf bytes.Buffer
	retryCmd.Stdout = &stdoutBuf
	retryCmd.Stderr = &stderrBuf

	if err := retryCmd.Run(); err != nil {
		_ = os.RemoveAll(stagingDir)
		return fmt.Errorf("patch failed even after cleanup: %w (original: %v)", err, originalErr)
	}

	logger.Info("Patch applied successfully after cleanup")
	cleanup(retryCmd, stagingDir, reporter)
	return nil
}

func cleanup(cmd *exec.Cmd, stagingDir string, reporter *progress.Reporter) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
	}
	_ = os.RemoveAll(stagingDir)
	if reporter != nil {
		reporter.Report(progress.StagePatch, 80, "Game patched!")
	}
}

func writeDebugLog(args []string, gameDir, stdout, stderr string, err error) string {
	debugLogPath := filepath.Join(env.GetCacheDir(), fmt.Sprintf("butler-debug-%d.log", time.Now().Unix()))
	content := fmt.Sprintf(
		"Time: %s\nCommand: %v\nGame Dir: %s\n\nSTDOUT:\n%s\n\nSTDERR:\n%s\n\nError: %v\n",
		time.Now().Format(time.RFC3339),
		args,
		gameDir,
		stdout,
		stderr,
		err,
	)
	_ = os.WriteFile(debugLogPath, []byte(content), 0644)
	return debugLogPath
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

	if _, pwrErr := os.Stat(pwrDest); pwrErr == nil {
		if _, sigErr := os.Stat(sigDest); sigErr == nil {
			if reporter != nil {
				reporter.Report(progress.StagePWR, 100, "Patch files cached")
			}
			return pwrDest, sigDest, nil
		}
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

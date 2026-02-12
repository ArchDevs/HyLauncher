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
	"runtime"
	"strconv"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/platform"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
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
	var pwrPath string

	// Fetch patch steps from API
	steps, err := fetchPatchSteps(ctx, branch, currentVer)
	if err != nil {
		return fmt.Errorf("fetch patch steps: %w", err)
	}

	if len(steps) == 0 {
		return fmt.Errorf("no patch steps available")
	}

	// Apply each patch step
	for i, step := range steps {
		// Stop if we've reached the target version
		if targetVer > 0 && step.From >= targetVer {
			break
		}

		if reporter != nil {
			reporter.Report(progress.StagePatch, 0, fmt.Sprintf("Patching %d → %d (%d/%d)", step.From, step.To, i+1, len(steps)))
		}

		// Download PWR and signature files for this step
		pwrPath, sigPath, err := downloadPatchStep(ctx, step, reporter)
		if err != nil {
			return fmt.Errorf("download patch step %d→%d: %w", step.From, step.To, err)
		}

		// Apply the patch to the specified version directory
		if err := applyPWR(ctx, pwrPath, sigPath, branch, versionDir, reporter); err != nil {
			return fmt.Errorf("apply patch %d→%d: %w", step.From, step.To, err)
		}
	}

	_ = os.RemoveAll(pwrPath)

	return nil
}

func applyPWR(ctx context.Context, pwrFile string, sigFile string, branch string, version string, reporter *progress.Reporter) error {
	gameDir := env.GetGameDir(branch, version)
	// Keep staging dir OUTSIDE game directory to avoid verification issues
	stagingDir := filepath.Join(env.GetCacheDir(), "staging-temp")

	// Clean up any previous incomplete staging state
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if reporter != nil {
		reporter.Report(progress.StagePatch, 60, "Applying game patch...")
	}

	if err := cmd.Run(); err != nil {
		// Clean up staging on failure
		_ = os.RemoveAll(stagingDir)
		return err
	}

	if cmd.Process != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Process.Release()
	}

	// Clean up staging directory
	_ = os.RemoveAll(stagingDir)

	if reporter != nil {
		reporter.Report(progress.StagePatch, 80, "Game patched!")
	}
	return nil
}

func fetchPatchSteps(ctx context.Context, branch string, currentVer int) ([]PatchStep, error) {
	reqBody := PatchRequest{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
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

	// Check if files are already cached
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

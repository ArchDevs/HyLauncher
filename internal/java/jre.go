package java

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/download"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/logger"
)

var (
	ErrJavaNotFound = fmt.Errorf("java not found")
	ErrJavaBroken   = fmt.Errorf("java broken")
)

type JREPlatform struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type JREJSON struct {
	Version     string                            `json:"version"`
	DownloadURL map[string]map[string]JREPlatform `json:"download_url"`
}

func GetJREVersionDir(version string) string {
	return filepath.Join(env.GetJREDir(), version)
}

func FetchJREManifest(branch string) (*JREJSON, error) {
	url := fmt.Sprintf("https://launcher.hytale.com/version/%s/jre.json", branch)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var jreData JREJSON
	if err := json.NewDecoder(resp.Body).Decode(&jreData); err != nil {
		return nil, err
	}
	return &jreData, nil
}

func verifyJREVersion(version string) error {
	javaBin := getJavaExecutablePathForVersion(version)

	if !fileutil.FileExistsNative(javaBin) {
		return ErrJavaNotFound
	}

	if !fileutil.FileFunctional(javaBin) {
		return ErrJavaBroken
	}

	return nil
}

func EnsureJRE(ctx context.Context, branch string, reporter *progress.Reporter) error {
	logger.Info("Checking JRE", "branch", branch)

	manifest, err := FetchJREManifest(branch)
	if err != nil {
		logger.Error("Failed to fetch JRE manifest", "branch", branch, "error", err)
		return err
	}

	jreVersion := manifest.Version
	jreDir := GetJREVersionDir(jreVersion)

	logger.Info("JRE version required", "version", jreVersion)

	if verifyJREVersion(jreVersion) == nil {
		logger.Info("JRE already installed", "version", jreVersion)
		if reporter != nil {
			reporter.Report(progress.StageJRE, 100, fmt.Sprintf("JRE %s ready", jreVersion))
		}
		return nil
	}

	if reporter != nil {
		reporter.Report(progress.StageJRE, 0, fmt.Sprintf("Installing JRE %s", jreVersion))
	}

	osName := env.GetOS()
	arch := env.GetArch()
	cacheDir := env.GetCacheDir()

	logger.Info("Installing JRE", "version", jreVersion, "os", osName, "arch", arch)

	if err := downloadAndInstallJRE(ctx, manifest, jreDir, cacheDir, osName, arch, reporter); err != nil {
		logger.Error("Failed to install JRE", "version", jreVersion, "error", err)
		_ = os.RemoveAll(jreDir)
		return err
	}

	logger.Info("JRE installed successfully", "version", jreVersion)

	if reporter != nil {
		reporter.Report(progress.StageJRE, 100, fmt.Sprintf("JRE %s installed", jreVersion))
	}
	return nil
}

func downloadAndInstallJRE(ctx context.Context, manifest *JREJSON, jreDir, cacheDir, osName, arch string, reporter *progress.Reporter) error {
	osData, ok := manifest.DownloadURL[osName]
	if !ok {
		return fmt.Errorf("no JRE for OS: %s", osName)
	}

	platform, ok := osData[arch]
	if !ok {
		return fmt.Errorf("no JRE for arch: %s on %s", arch, osName)
	}

	fileName := filepath.Base(platform.URL)
	cacheFile := filepath.Join(cacheDir, fileName)

	_ = os.MkdirAll(cacheDir, 0755)
	_ = os.MkdirAll(filepath.Dir(jreDir), 0755)

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		logger.Info("Downloading JRE", "url", platform.URL, "file", fileName)
		scaler := progress.NewScaler(reporter, progress.StageJRE, 0, 90)
		if err := download.DownloadWithReporter(ctx, cacheFile, platform.URL, fileName, reporter, progress.StageJRE, scaler); err != nil {
			logger.Error("Failed to download JRE", "url", platform.URL, "error", err)
			_ = os.Remove(cacheFile)
			return err
		}
		logger.Info("JRE downloaded", "file", fileName)
	} else {
		logger.Info("JRE archive cached", "file", cacheFile)
		if reporter != nil {
			reporter.Report(progress.StageJRE, 90, "JRE archive cached")
		}
	}

	if reporter != nil {
		reporter.Report(progress.StageJRE, 92, "Verifying JRE integrity")
	}
	logger.Info("Verifying JRE integrity", "file", cacheFile)
	if err := fileutil.VerifySHA256(cacheFile, platform.SHA256); err != nil {
		logger.Error("JRE integrity check failed", "file", cacheFile, "error", err)
		_ = os.Remove(cacheFile)
		return err
	}
	logger.Info("JRE integrity verified")

	tempDir := jreDir + ".tmp"
	_ = os.RemoveAll(tempDir)

	if reporter != nil {
		reporter.Report(progress.StageJRE, 95, "Extracting JRE")
	}
	logger.Info("Extracting JRE", "archive", cacheFile, "dest", tempDir)

	if err := extractJRE(cacheFile, tempDir); err != nil {
		logger.Error("Failed to extract JRE", "error", err)
		_ = os.RemoveAll(tempDir)
		return err
	}

	if err := flattenJREDir(tempDir); err != nil {
		logger.Error("Failed to flatten JRE directory", "error", err)
		_ = os.RemoveAll(tempDir)
		return err
	}

	if reporter != nil {
		reporter.Report(progress.StageJRE, 98, "Finalizing JRE installation")
	}

	_ = os.RemoveAll(jreDir)

	var finalErr error
	for i := 0; i < 5; i++ {
		finalErr = os.Rename(tempDir, jreDir)
		if finalErr == nil {
			logger.Info("JRE installation finalized", "dir", jreDir)
			break
		}
		logger.Warn("Failed to rename JRE directory, retrying", "attempt", i+1, "error", finalErr)
		time.Sleep(2 * time.Second)
	}
	if finalErr != nil {
		logger.Error("Failed to finalize JRE installation after retries", "error", finalErr)
		return fmt.Errorf("failed to finalize JRE installation: %w", finalErr)
	}

	if runtime.GOOS != "windows" {
		javaExec := getJavaExecutablePathForVersion(manifest.Version)
		_ = os.Chmod(javaExec, 0755)
	}

	_ = os.Remove(cacheFile)
	return nil
}

func VerifyJRE(branch string) error {
	manifest, err := FetchJREManifest(branch)
	if err != nil {
		return err
	}
	return verifyJREVersion(manifest.Version)
}

func GetJavaExec(branch string) (string, error) {
	manifest, err := FetchJREManifest(branch)
	if err != nil {
		return "", err
	}

	if err := verifyJREVersion(manifest.Version); err != nil {
		return "", err
	}

	return getJavaExecutablePathForVersion(manifest.Version), nil
}

func getJavaExecutablePathForVersion(version string) string {
	base := GetJREVersionDir(version)
	if runtime.GOOS == "darwin" {
		return filepath.Join(base, "Contents", "Home", "bin", "java")
	} else if runtime.GOOS == "windows" {
		return filepath.Join(base, "bin", "java.exe")
	}
	return filepath.Join(base, "bin", "java")
}

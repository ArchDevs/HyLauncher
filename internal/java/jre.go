package java

import (
	"context"
	"encoding/json"
	"errors"
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

func EnsureJRE(ctx context.Context, reporter *progress.Reporter) error {
	osName := env.GetOS()
	arch := env.GetArch()
	basePath := env.GetDefaultAppDir()

	cacheDir := filepath.Join(basePath, "cache")
	jreDir := filepath.Join(basePath, "release", "package", "jre")
	latestDir := filepath.Join(jreDir, "latest")

	err := VerifyJRE()
	if err != nil {
		if errors.Is(err, ErrJavaBroken) || errors.Is(err, ErrJavaNotFound) {
			if reinstallErr := ReinstallJRE(jreDir, latestDir, cacheDir, osName, arch, reporter); reinstallErr != nil {
				return reinstallErr
			}
		} else {
			return err
		}
	}
	return nil
}

func DownloadJRE(jreDir, latestDir, cacheDir, osName, arch string, reporter *progress.Reporter) error {
	resp, err := http.Get("https://launcher.hytale.com/version/release/jre.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jreData JREJSON
	if err := json.NewDecoder(resp.Body).Decode(&jreData); err != nil {
		return err
	}

	osData, ok := jreData.DownloadURL[osName]
	if !ok {
		return fmt.Errorf("no JRE for OS: %s", osName)
	}

	platform, ok := osData[arch]
	if !ok {
		return fmt.Errorf("no JRE for arch: %s on %s", arch, osName)
	}

	fileName := filepath.Base(platform.URL)
	cacheFile := filepath.Join(cacheDir, fileName)
	tempCacheFile := cacheFile + ".tmp"

	_ = os.MkdirAll(cacheDir, 0755)
	_ = os.MkdirAll(jreDir, 0755)

	// Download JRE archive if not cached
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		_ = os.Remove(tempCacheFile)

		// Create a scaler for the download portion (0-90%)
		scaler := progress.NewScaler(reporter, progress.StageJRE, 0, 90)

		if err := download.DownloadWithReporter(cacheFile, platform.URL, fileName, reporter, progress.StageJRE, scaler); err != nil {
			_ = os.Remove(tempCacheFile)
			return err
		}
	} else {
		reporter.Report(progress.StageJRE, 90, "JRE archive cached")
	}

	// Verify hash sha256
	reporter.Report(progress.StageJRE, 92, "Verifying JRE integrity")
	if err := fileutil.VerifySHA256(cacheFile, platform.SHA256); err != nil {
		_ = os.Remove(cacheFile)
		return err
	}

	// Extract into temporary folder
	tempDir := filepath.Join(jreDir, "tmp-"+jreData.Version)
	_ = os.RemoveAll(tempDir)

	reporter.Report(progress.StageJRE, 95, "Extracting JRE")
	if err := extractJRE(cacheFile, tempDir); err != nil {
		return err
	}

	// Flatten directory if needed
	if err := flattenJREDir(tempDir); err != nil {
		return err
	}

	// Atomic rename: tmp -> latest
	reporter.Report(progress.StageJRE, 98, "Finalizing JRE installation...")

	_ = os.RemoveAll(latestDir)

	// On Windows, retry a few times because antivirus may lock files
	for i := 0; i < 5; i++ {
		err = os.Rename(tempDir, latestDir)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to finalize JRE installation: %w", err)
	}

	// Ensure java binary is executable (Linux/macOS)
	if runtime.GOOS != "windows" {
		javaExec := filepath.Join(latestDir, "bin", "java")
		_ = os.Chmod(javaExec, 0755)
	}

	// Cleanup cache
	_ = os.Remove(cacheFile)
	return nil
}

func ReinstallJRE(jreDir, latestDir, cacheDir, osName, arch string, reporter *progress.Reporter) error {
	err := os.RemoveAll(jreDir)
	if err != nil {
		fmt.Println("Warning: can not delete jre folder")
		return err
	}
	reporter.Report(progress.StageJRE, 0, "Starting JRE installation")
	err = os.MkdirAll(jreDir, 0755)
	if err != nil {
		fmt.Println("Warning: can create jre folder")
		return err
	}

	err = DownloadJRE(jreDir, latestDir, cacheDir, osName, arch, reporter)
	if err != nil {
		fmt.Println("Warning: can not download jre")
		return err
	}

	reporter.Report(progress.StageJRE, 100, "JRE installed successfully")
	return nil
}

func VerifyJRE() error {
	jreDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "jre", "latest")
	javaBin := filepath.Join(jreDir, "bin", "java")

	if !fileutil.FileExistsNative(javaBin) {
		fmt.Println("Warning: JRE not found, fallback to system java")
		return ErrJavaNotFound
	}

	if !fileutil.FileFunctional(javaBin) {
		fmt.Println("Warning: java executable is broken")
		return ErrJavaBroken
	}

	return nil
}

func GetJavaExec() (string, error) {
	err := VerifyJRE()
	if err != nil {
		return "", err
	}

	javaBin := filepath.Join(
		env.GetDefaultAppDir(),
		"release",
		"package",
		"jre",
		"latest",
		"bin",
		"java",
	)
	if runtime.GOOS == "windows" {
		javaBin += ".exe"
	}
	return javaBin, nil
}

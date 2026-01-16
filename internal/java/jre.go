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
	"HyLauncher/internal/models/hyerrors"
	"HyLauncher/internal/util"
	"HyLauncher/internal/util/download"
)

type JREPlatform struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type JREJSON struct {
	Version     string                            `json:"version"`
	DownloadURL map[string]map[string]JREPlatform `json:"download_url"`
}

func DownloadJRE(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	osName := env.GetOS()
	arch := env.GetArch()
	basePath := env.GetDefaultAppDir()

	cacheDir := filepath.Join(basePath, "cache")
	jreDir := filepath.Join(basePath, "release", "package", "jre")
	latestDir := filepath.Join(jreDir, "latest")

	if isJREInstalled(latestDir) {
		if progressCallback != nil {
			progressCallback("jre", 100, "JRE already installed", "", "", 0, 0)
		}
		return nil
	}

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
		if err := download.DownloadWithProgress(tempCacheFile, platform.URL, "jre", 0.9, progressCallback); err != nil {
			_ = os.Remove(tempCacheFile)
			return err
		}

		if err := os.Rename(tempCacheFile, cacheFile); err != nil {
			_ = os.Remove(tempCacheFile)
			return err
		}
	}

	// Verify hash sha256
	if progressCallback != nil {
		progressCallback("jre", 92, "Verifying JRE integrity...", fileName, "", 0, 0)
	}
	if err := util.VerifySHA256(cacheFile, platform.SHA256); err != nil {
		_ = os.Remove(cacheFile)
		return err
	}

	// Extract into temporary folder
	tempDir := filepath.Join(jreDir, "tmp-"+jreData.Version)
	_ = os.RemoveAll(tempDir)

	if progressCallback != nil {
		progressCallback("jre", 95, "Extracting JRE...", fileName, "", 0, 0)
	}
	if err := extractJRE(cacheFile, tempDir); err != nil {
		return err
	}

	// Flatten directory if needed
	if err := flattenJREDir(tempDir); err != nil {
		return err
	}

	// Atomic rename: tmp -> latest
	if progressCallback != nil {
		progressCallback("jre", 98, "Finalizing JRE installation...", fileName, "", 0, 0)
	}

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

	if progressCallback != nil {
		progressCallback("jre", 100, "JRE installed successfully", fileName, "", 0, 0)
	}

	return nil
}

func GetJavaExec() (string, error) {
	jreDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "jre", "latest")
	javaBin := filepath.Join(jreDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaBin += ".exe"
	}

	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		fmt.Println("Warning: JRE not found, fallback to system java")
		return "", hyerrors.ErrJavaNotFound
	}

	if ok := isJavaFunctional(javaBin); ok == false {
		return "", hyerrors.ErrJavaBroken
	}

	return javaBin, nil
}

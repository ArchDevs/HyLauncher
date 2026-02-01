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
	manifest, err := FetchJREManifest(branch)
	if err != nil {
		return err
	}

	jreVersion := manifest.Version
	jreDir := GetJREVersionDir(jreVersion)

	if verifyJREVersion(jreVersion) == nil {
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

	if err := downloadAndInstallJRE(ctx, manifest, jreDir, cacheDir, osName, arch, reporter); err != nil {
		_ = os.RemoveAll(jreDir)
		return err
	}

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
		scaler := progress.NewScaler(reporter, progress.StageJRE, 0, 90)
		if err := download.DownloadWithReporter(ctx, cacheFile, platform.URL, fileName, reporter, progress.StageJRE, scaler); err != nil {
			_ = os.Remove(cacheFile)
			return err
		}
	} else if reporter != nil {
		reporter.Report(progress.StageJRE, 90, "JRE archive cached")
	}

	if reporter != nil {
		reporter.Report(progress.StageJRE, 92, "Verifying JRE integrity")
	}
	if err := fileutil.VerifySHA256(cacheFile, platform.SHA256); err != nil {
		_ = os.Remove(cacheFile)
		return err
	}

	tempDir := jreDir + ".tmp"
	_ = os.RemoveAll(tempDir)

	if reporter != nil {
		reporter.Report(progress.StageJRE, 95, "Extracting JRE")
	}

	if err := extractJRE(cacheFile, tempDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return err
	}

	if err := flattenJREDir(tempDir); err != nil {
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
			break
		}
		time.Sleep(2 * time.Second)
	}
	if finalErr != nil {
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

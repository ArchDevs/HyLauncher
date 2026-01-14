package java

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyLauncher/internal/env"
)

type JREPlatform struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type JREJSON struct {
	Version     string                            `json:"version"`
	DownloadURL map[string]map[string]JREPlatform `json:"download_url"`
}

type progressReader struct {
	reader    io.Reader
	total     int64
	read      int64
	lastPrint time.Time
	callback  func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)
	fileName  string
	startTime time.Time
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.reader.Read(b)
	p.read += int64(n)

	if time.Since(p.lastPrint) > 200*time.Millisecond {
		percent := float64(p.read) / float64(p.total) * 100
		elapsed := time.Since(p.startTime).Seconds()
		speed := ""
		if elapsed > 0 {
			mbps := float64(p.read) / 1024 / 1024 / elapsed
			speed = fmt.Sprintf("%.2f MB/s", mbps)
		}

		if p.callback != nil {
			p.callback("jre", percent, "Downloading JRE...", p.fileName, speed, p.read, p.total)
		}

		fmt.Printf("\rDownloading... %.1f%%", percent)
		p.lastPrint = time.Now()
	}
	return n, err
}

func DownloadJRE(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	osName := env.GetOS()
	arch := env.GetArch()

	basePath := env.GetDefaultAppDir()

	cacheDir := filepath.Join(basePath, "cache")
	jreLatest := filepath.Join(basePath, "release", "package", "jre", "latest")

	if isJREInstalled(jreLatest) {
		fmt.Println("JRE already installed, skipping")
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

	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		fmt.Println("Downloading JRE...")
		if progressCallback != nil {
			progressCallback("jre", 0, "Downloading JRE...", fileName, "", 0, 0)
		}

		resp2, err := http.Get(platform.URL)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()

		out, err := os.Create(cacheFile)
		if err != nil {
			return err
		}
		defer out.Close()

		pr := &progressReader{
			reader:    resp2.Body,
			total:     resp2.ContentLength,
			callback:  progressCallback,
			fileName:  fileName,
			startTime: time.Now(),
		}

		if _, err := io.Copy(out, pr); err != nil {
			return err
		}

		fmt.Println("\nDownload complete")
	}

	fmt.Println("Verifying JRE...")
	if progressCallback != nil {
		progressCallback("jre", 90, "Verifying JRE...", fileName, "", 0, 0)
	}

	if err := verifySHA256(cacheFile, platform.SHA256); err != nil {
		_ = os.Remove(cacheFile)
		return err
	}

	fmt.Println("Extracting JRE...")
	if progressCallback != nil {
		progressCallback("jre", 95, "Extracting JRE...", fileName, "", 0, 0)
	}

	if err := extractJRE(cacheFile, jreLatest); err != nil {
		return err
	}

	if osName != "windows" {
		java := filepath.Join(jreLatest, "bin", "java")
		_ = os.Chmod(java, 0755)
	}

	flattenJREDir(jreLatest)

	if err := os.Remove(cacheFile); err != nil {
		fmt.Println("Warning: failed to remove cached JRE:", err)
	}

	fmt.Println("JRE installed successfully")
	if progressCallback != nil {
		progressCallback("jre", 100, "JRE installed", "", "", 0, 0)
	}

	return nil
}

func GetJavaExec() string {
	jreDir := filepath.Join(env.GetDefaultAppDir(), "release", "package", "jre", "latest")
	javaBin := filepath.Join(jreDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaBin += ".exe"
	}

	// Check if it exists
	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		// If java is missing, just return "java" and hope it's in PATH (fallback)
		fmt.Println("Warning: JRE not found, fallback to system java")
		return "java"
	}

	return javaBin
}

func verifySHA256(filePath, expected string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	if sum != expected {
		return fmt.Errorf("SHA256 mismatch: expected %s got %s", expected, sum)
	}
	return nil
}

func isJREInstalled(jreDir string) bool {
	java := filepath.Join(jreDir, "bin", "java")
	if runtime.GOOS == "windows" {
		java += ".exe"
	}
	_, err := os.Stat(java)
	return err == nil
}

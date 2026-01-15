package pwr

import (
	"HyLauncher/internal/env"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type VersionInfo struct {
	Version int `json:"version"`
}

type VersionCheckResult struct {
	LatestVersion int
	Error         error
	CheckedURLs   []string
	SuccessURL    string
}

func GetLocalVersion() string {
	path := filepath.Join(env.GetDefaultAppDir(), "release", "version.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "0"
	}

	var info VersionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return "0"
	}

	return strconv.Itoa(info.Version)
}

func SaveLocalVersion(v int) error {
	path := filepath.Join(env.GetDefaultAppDir(), "release", "version.json")
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.Marshal(VersionInfo{Version: v})
	return os.WriteFile(path, data, 0644)
}

// FindLatestVersion discovers the newest version using parallel HEAD requests
func FindLatestVersion(versionType string) int {
	result := FindLatestVersionWithDetails(versionType)

	if result.Error != nil {
		fmt.Printf("Error finding latest version: %v\n", result.Error)
		fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Checked %d URLs\n", len(result.CheckedURLs))
		if len(result.CheckedURLs) > 0 {
			fmt.Printf("Sample URL: %s\n", result.CheckedURLs[0])
		}
	}

	return result.LatestVersion
}

// FindLatestVersionWithDetails returns detailed information about the version check
func FindLatestVersionWithDetails(versionType string) VersionCheckResult {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	result := VersionCheckResult{
		LatestVersion: 0,
		CheckedURLs:   make([]string, 0),
	}

	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 5 * time.Second, // Increased from 2s to 5s
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Try known good versions first for faster startup
	knownVersions := []int{100, 50, 25, 10, 5, 1}
	for _, v := range knownVersions {
		url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
			osName, arch, versionType, v)

		result.CheckedURLs = append(result.CheckedURLs, url)

		resp, err := client.Head(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			result.LatestVersion = v
			result.SuccessURL = url
			fmt.Printf("Found version %d, searching for latest...\n", v)
			break
		}
	}

	// If no known version worked, we have a problem
	if result.LatestVersion == 0 {
		result.Error = fmt.Errorf(
			"cannot reach game server or no versions available for %s/%s. "+
				"Please check:\n"+
				"1. Internet connection\n"+
				"2. Firewall/antivirus settings\n"+
				"3. Game server status\n"+
				"Platform: %s %s",
			osName, arch, osName, arch,
		)
		return result
	}

	// Now do binary search to find the actual latest
	latestFound := result.LatestVersion
	batchSize := 10
	maxVersion := 500

	// Start from the last known good version
	start := latestFound + 1

	for start < maxVersion {
		var wg sync.WaitGroup
		batchResults := make(chan int, batchSize)
		end := start + batchSize
		if end > maxVersion {
			end = maxVersion
		}

		for i := start; i < end; i++ {
			wg.Add(1)
			go func(v int) {
				defer wg.Done()

				url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
					osName, arch, versionType, v)

				resp, err := client.Head(url)
				if err == nil && resp.StatusCode == http.StatusOK {
					batchResults <- v
				} else {
					batchResults <- 0
				}
			}(i)
		}

		wg.Wait()
		close(batchResults)

		maxInBatch := 0
		for v := range batchResults {
			if v > maxInBatch {
				maxInBatch = v
			}
		}

		if maxInBatch > latestFound {
			latestFound = maxInBatch
			result.LatestVersion = latestFound
			result.SuccessURL = fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
				osName, arch, versionType, latestFound)
			start = latestFound + 1
		} else {
			// No higher version found, we're done
			break
		}
	}

	fmt.Printf("Latest version found: %d\n", result.LatestVersion)
	return result
}

// VerifyVersionExists checks if a specific version exists on the server
func VerifyVersionExists(versionType string, version int) error {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
		osName, arch, versionType, version)

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Head(url)
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("version %d not found (HTTP %d)", version, resp.StatusCode)
	}

	return nil
}

// TestConnection tests if we can reach the game server at all
func TestConnection() error {
	testURL := "https://game-patches.hytale.com/"

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head(testURL)
	if err != nil {
		return fmt.Errorf("cannot reach game server: %w", err)
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("game server error (HTTP %d)", resp.StatusCode)
	}

	return nil
}

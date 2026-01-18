package patch

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

// TODO FULL REFACTOR

type VersionInfo struct {
	Version int `json:"version"`
}

type VersionCheckResult struct {
	LatestVersion int
	Error         error
	CheckedURLs   []string
	SuccessURL    string
}

var (
	versionCache      = make(map[string]*VersionCheckResult)
	versionCacheMutex sync.RWMutex
	versionCacheTTL   = 5 * time.Minute
	lastCheckTime     = make(map[string]time.Time)
)

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

func FindLatestVersionWithDetails(versionType string) VersionCheckResult {
	cacheKey := fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, versionType)

	// Check cache first
	versionCacheMutex.RLock()
	if cached, exists := versionCache[cacheKey]; exists {
		if time.Since(lastCheckTime[cacheKey]) < versionCacheTTL {
			fmt.Printf("Using cached version: %d\n", cached.LatestVersion)
			versionCacheMutex.RUnlock()
			return *cached
		}
	}
	versionCacheMutex.RUnlock()

	fmt.Println("Performing version check...")
	result := performVersionCheck(versionType)

	// Cache the result
	versionCacheMutex.Lock()
	versionCache[cacheKey] = &result
	lastCheckTime[cacheKey] = time.Now()
	versionCacheMutex.Unlock()

	return result
}

func ClearVersionCache() {
	versionCacheMutex.Lock()
	versionCache = make(map[string]*VersionCheckResult)
	lastCheckTime = make(map[string]time.Time)
	versionCacheMutex.Unlock()
	fmt.Println("Version cache cleared")
}

func performVersionCheck(versionType string) VersionCheckResult {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	result := VersionCheckResult{
		LatestVersion: 0,
		CheckedURLs:   make([]string, 0),
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Try known good versions first to establish a baseline
	knownVersions := []int{100, 50, 25, 10, 5, 1}
	var foundBase int

	fmt.Println("Searching for base version...")
	for _, v := range knownVersions {
		url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
			osName, arch, versionType, v)

		result.CheckedURLs = append(result.CheckedURLs, url)

		resp, err := client.Head(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			foundBase = v
			result.LatestVersion = v
			result.SuccessURL = url
			fmt.Printf("Found base version %d\n", v)
			break
		}

		// Small delay to avoid overwhelming network
		time.Sleep(200 * time.Millisecond)
	}

	// If no known version worked, we have a problem
	if foundBase == 0 {
		result.Error = fmt.Errorf(
			"cannot reach game server or no versions available for %s/%s",
			osName, arch,
		)
		return result
	}

	// For small base versions (≤10), just do linear search up to a reasonable max
	if foundBase <= 10 {
		fmt.Println("Doing linear search for latest version...")
		maxCheck := foundBase + 50
		if maxCheck > 200 {
			maxCheck = 200
		}

		for v := foundBase + 1; v <= maxCheck; v++ {
			url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
				osName, arch, versionType, v)

			resp, err := client.Head(url)
			time.Sleep(200 * time.Millisecond)

			if err == nil && resp.StatusCode == http.StatusOK {
				result.LatestVersion = v
				result.SuccessURL = url
				fmt.Printf("Found version %d\n", v)
			} else {
				break
			}
		}

		fmt.Printf("Latest version found: %d\n", result.LatestVersion)
		return result
	}

	// For larger base versions, use exponential search
	fmt.Println("Searching for latest version...")
	step := foundBase
	current := foundBase
	maxVersion := 500

	for current < maxVersion {
		next := current + step
		if next > maxVersion {
			next = maxVersion
		}

		url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
			osName, arch, versionType, next)

		resp, err := client.Head(url)
		time.Sleep(200 * time.Millisecond)

		if err == nil && resp.StatusCode == http.StatusOK {
			result.LatestVersion = next
			result.SuccessURL = url
			current = next
			step *= 2
			fmt.Printf("Version %d exists, trying higher...\n", next)
		} else {
			break
		}
	}

	// Binary search between last known good and failed version
	low := result.LatestVersion
	high := current + step
	if high > maxVersion {
		high = maxVersion
	}

	if high > low {
		fmt.Printf("Binary search between %d and %d...\n", low, high)

		for low < high {
			mid := (low + high + 1) / 2

			url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
				osName, arch, versionType, mid)

			resp, err := client.Head(url)
			time.Sleep(200 * time.Millisecond)

			if err == nil && resp.StatusCode == http.StatusOK {
				result.LatestVersion = mid
				result.SuccessURL = url
				low = mid
				fmt.Printf("Version %d exists\n", mid)
			} else {
				high = mid - 1
			}
		}
	}

	fmt.Printf("Latest version found: %d\n", result.LatestVersion)
	return result
}

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

func TestConnection(testURL string) error {
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

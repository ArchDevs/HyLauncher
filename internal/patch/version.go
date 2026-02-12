package patch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type VersionCheckResult struct {
	LatestVersion int
	Error         error
}

type AllVersionsResult struct {
	Versions []int
	Error    error
}

type cache struct {
	mu            sync.RWMutex
	latestVersion map[string]*VersionCheckResult
	versionExists map[string]bool
	allVersions   map[string]*AllVersionsResult
	lastSet       map[string]time.Time
	ttl           time.Duration
}

var versionCache = &cache{
	latestVersion: make(map[string]*VersionCheckResult),
	versionExists: make(map[string]bool),
	allVersions:   make(map[string]*AllVersionsResult),
	lastSet:       make(map[string]time.Time),
	ttl:           5 * time.Minute,
}

func (c *cache) getLatest(key string) (*VersionCheckResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, exists := c.latestVersion[key]
	if !exists || time.Since(c.lastSet[key]) >= c.ttl {
		return nil, false
	}
	return result, true
}

func (c *cache) setLatest(key string, result *VersionCheckResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.latestVersion[key] = result
	c.lastSet[key] = time.Now()
}

func (c *cache) checkVersion(key string) (bool, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	exists, cached := c.versionExists[key]
	return exists, cached
}

func (c *cache) setVersion(key string, exists bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.versionExists[key] = exists
}

func (c *cache) getAllVersions(key string) (*AllVersionsResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, exists := c.allVersions[key]
	if !exists || time.Since(c.lastSet[key]) >= c.ttl {
		return nil, false
	}
	return result, true
}

func (c *cache) setAllVersions(key string, result *AllVersionsResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.allVersions[key] = result
	c.lastSet[key] = time.Now()
}

func (c *cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.latestVersion = make(map[string]*VersionCheckResult)
	c.versionExists = make(map[string]bool)
	c.allVersions = make(map[string]*AllVersionsResult)
	c.lastSet = make(map[string]time.Time)
}

func FindLatestVersion(branch string) (int, error) {
	key := cacheKey(branch)

	if cached, ok := versionCache.getLatest(key); ok {
		return cached.LatestVersion, cached.Error
	}

	result := findLatestVersion(branch)
	versionCache.setLatest(key, &result)

	return result.LatestVersion, result.Error
}

func ListAllVersions(branch string) ([]int, error) {
	key := cacheKey(branch)

	if cached, ok := versionCache.getAllVersions(key); ok {
		return cached.Versions, cached.Error
	}

	result := listAllVersions(branch)
	versionCache.setAllVersions(key, &result)

	return result.Versions, result.Error
}

func ClearVersionCache() {
	versionCache.clear()
}

func VerifyVersionExists(branch string, version int) error {
	client := createClient()

	if exists := checkVersionExists(client, branch, version); exists {
		return nil
	}

	return fmt.Errorf("version %d not found", version)
}

func findLatestVersion(branch string) VersionCheckResult {
	// Call API with version 1 to get all steps from 1 to latest
	steps, err := fetchPatchStepsFromAPI(branch, 1)
	if err != nil {
		return VersionCheckResult{
			Error: fmt.Errorf("cannot reach API or no patches available for %s/%s: %w", runtime.GOOS, runtime.GOARCH, err),
		}
	}

	if len(steps) == 0 {
		return VersionCheckResult{
			Error: fmt.Errorf("no patches available for %s/%s", runtime.GOOS, runtime.GOARCH),
		}
	}

	// The latest version is the "to" field of the last step
	latest := steps[len(steps)-1].To
	return VersionCheckResult{LatestVersion: latest}
}

func listAllVersions(branch string) AllVersionsResult {
	// Call API with version 1 to get all steps from 1 to latest
	steps, err := fetchPatchStepsFromAPI(branch, 1)
	if err != nil {
		return AllVersionsResult{
			Error: fmt.Errorf("cannot reach API or no patches available for %s/%s: %w", runtime.GOOS, runtime.GOARCH, err),
		}
	}

	if len(steps) == 0 {
		return AllVersionsResult{
			Error: fmt.Errorf("no patches available for %s/%s", runtime.GOOS, runtime.GOARCH),
		}
	}

	// Extract all unique versions from steps
	versionMap := make(map[int]bool)
	for _, step := range steps {
		versionMap[step.From] = true
		versionMap[step.To] = true
	}

	// Convert to sorted slice
	var versions []int
	for v := range versionMap {
		versions = append(versions, v)
	}

	// Sort in ascending order
	for i := 0; i < len(versions); i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i] > versions[j] {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	return AllVersionsResult{Versions: versions}
}

// fetchPatchStepsFromAPI calls the PWR API to get patch steps
func fetchPatchStepsFromAPI(branch string, currentVer int) ([]PatchStep, error) {
	reqBody := PatchRequest{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Branch:  branch,
		Version: fmt.Sprintf("%d", currentVer),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("GET", "https://api.hylauncher.fun/v1/pwr", bytes.NewBuffer(jsonData))
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

func checkVersionExists(client *http.Client, branch string, version int) bool {
	key := versionCacheKey(branch, version)

	if exists, cached := versionCache.checkVersion(key); cached {
		return exists
	}

	url := buildPatchURL(branch, version)
	resp, err := client.Head(url)
	if resp != nil {
		defer resp.Body.Close()
	}

	exists := err == nil && resp.StatusCode == http.StatusOK
	versionCache.setVersion(key, exists)

	time.Sleep(100 * time.Millisecond)

	return exists
}

func createClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          3,
			MaxIdleConnsPerHost:   2,
			MaxConnsPerHost:       3,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func buildPatchURL(branch string, version int) string {
	return fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
		runtime.GOOS, runtime.GOARCH, branch, version)
}

func cacheKey(branch string) string {
	return fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, branch)
}

func versionCacheKey(branch string, version int) string {
	return fmt.Sprintf("%s-%d", cacheKey(branch), version)
}

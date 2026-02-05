package patch

import (
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

func ListAllVersionsBothBranches() (release []int, prerelease []int, err error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	wg.Add(2)

	go func() {
		defer wg.Done()
		versions, e := ListAllVersions("release")
		mu.Lock()
		release = versions
		if e != nil {
			errors = append(errors, fmt.Errorf("release: %w", e))
		}
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		versions, e := ListAllVersions("pre-release")
		mu.Lock()
		prerelease = versions
		if e != nil {
			errors = append(errors, fmt.Errorf("pre-release: %w", e))
		}
		mu.Unlock()
	}()

	wg.Wait()

	if len(errors) > 0 {
		return release, prerelease, fmt.Errorf("errors occurred: %v", errors)
	}

	return release, prerelease, nil
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
	client := createClient()

	base := findFirstVersion(client, branch)
	if base == 0 {
		return VersionCheckResult{
			Error: fmt.Errorf("cannot reach game servers or no patches available for %s/%s (check firewall/network)", runtime.GOOS, runtime.GOARCH),
		}
	}

	upper := findUpperBound(client, branch, base)
	latest := binarySearchLatest(client, branch, base, upper)

	return VersionCheckResult{LatestVersion: latest}
}

func listAllVersions(branch string) AllVersionsResult {
	client := createClient()

	base := findFirstVersion(client, branch)
	if base == 0 {
		return AllVersionsResult{
			Error: fmt.Errorf("cannot reach game servers or no patches available for %s/%s (check firewall/network)", runtime.GOOS, runtime.GOARCH),
		}
	}

	upper := findUpperBound(client, branch, base)
	versions := collectAllVersions(client, branch, base, upper)

	return AllVersionsResult{Versions: versions}
}

func findFirstVersion(client *http.Client, branch string) int {
	checkPoints := []int{1, 5, 10, 25}

	for _, v := range checkPoints {
		if checkVersionExists(client, branch, v) {
			return v
		}
	}

	return 0
}

func findUpperBound(client *http.Client, branch string, base int) int {
	current := base
	step := max(base, 10)

	for {
		next := current + step
		if checkVersionExists(client, branch, next) {
			current = next
			step *= 2
		} else {
			return next
		}

		if step > 1000 {
			break
		}
	}

	return current + step
}

func binarySearchLatest(client *http.Client, branch string, low, high int) int {
	latest := low

	for low < high {
		mid := (low + high + 1) / 2
		if checkVersionExists(client, branch, mid) {
			latest = mid
			low = mid
		} else {
			high = mid - 1
		}
	}

	return latest
}

func collectAllVersions(client *http.Client, branch string, start, end int) []int {
	var versions []int

	for v := start; v <= end; v++ {
		if checkVersionExists(client, branch, v) {
			versions = append(versions, v)
		}
	}

	return versions
}

func checkVersionExists(client *http.Client, branch string, version int) bool {
	key := versionCacheKey(branch, version)

	if exists, cached := versionCache.checkVersion(key); cached {
		return exists
	}

	url := buildPatchURL(branch, version)
	resp, err := client.Head(url)

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
			MaxIdleConns:          10,
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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

type cache struct {
	mu      sync.RWMutex
	data    map[string]*VersionCheckResult
	lastSet map[string]time.Time
	ttl     time.Duration
}

var versionCache = &cache{
	data:    make(map[string]*VersionCheckResult),
	lastSet: make(map[string]time.Time),
	ttl:     5 * time.Minute,
}

func (c *cache) get(key string) (*VersionCheckResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, exists := c.data[key]
	if !exists || time.Since(c.lastSet[key]) >= c.ttl {
		return nil, false
	}
	return result, true
}

func (c *cache) set(key string, result *VersionCheckResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = result
	c.lastSet[key] = time.Now()
}

func (c *cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*VersionCheckResult)
	c.lastSet = make(map[string]time.Time)
}

func FindLatestVersion(branch string) (int, error) {
	key := fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, branch)

	if cached, ok := versionCache.get(key); ok {
		return cached.LatestVersion, cached.Error
	}

	result := checkVersion(branch)
	versionCache.set(key, &result)

	return result.LatestVersion, result.Error
}

func ClearVersionCache() {
	versionCache.clear()
}

func checkVersion(branch string) VersionCheckResult {
	client := createClient()

	baseVersion := findBaseVersion(client, branch)
	if baseVersion == 0 {
		return VersionCheckResult{
			Error: fmt.Errorf("cannot reach game servers or no patches available for %s/%s (check firewall/network)", runtime.GOOS, runtime.GOARCH),
		}
	}

	latestVersion := findLatestVersion(client, branch, baseVersion)
	return VersionCheckResult{LatestVersion: latestVersion}
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

func findBaseVersion(client *http.Client, branch string) int {
	knownVersions := []int{25, 10, 5, 1}

	for _, v := range knownVersions {
		// Try each version with retry
		for attempt := 0; attempt < 2; attempt++ {
			if versionExists(client, branch, v) {
				return v
			}
			if attempt < 1 {
				time.Sleep(time.Second)
			}
		}
	}
	return 0
}

func findLatestVersion(client *http.Client, branch string, base int) int {
	if base <= 10 {
		return linearSearch(client, branch, base, min(base+50, 200))
	}

	upper := exponentialSearch(client, branch, base, 500)
	return binarySearch(client, branch, base, upper)
}

func linearSearch(client *http.Client, branch string, start, end int) int {
	latest := start
	for v := start + 1; v <= end; v++ {
		if versionExists(client, branch, v) {
			latest = v
		} else {
			break
		}
	}
	return latest
}

func exponentialSearch(client *http.Client, branch string, base, max int) int {
	current := base
	step := base

	for current < max {
		next := min(current+step, max)
		if versionExists(client, branch, next) {
			current = next
			step *= 2
		} else {
			return current + step
		}
	}
	return max
}

func binarySearch(client *http.Client, branch string, low, high int) int {
	latest := low

	for low < high {
		mid := (low + high + 1) / 2
		if versionExists(client, branch, mid) {
			latest = mid
			low = mid
		} else {
			high = mid - 1
		}
	}
	return latest
}

func versionExists(client *http.Client, branch string, version int) bool {
	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
		runtime.GOOS, runtime.GOARCH, branch, version)

	resp, err := client.Head(url)
	time.Sleep(200 * time.Millisecond)

	return err == nil && resp.StatusCode == http.StatusOK
}

// ListAvailableVersions returns all available game versions for the given branch.
// It reuses the same discovery strategy as FindLatestVersion, but then walks
// from the discovered base version up to the latest, checking which versions exist.
func ListAvailableVersions(branch string) ([]int, error) {
	client := createClient()

	baseVersion := findBaseVersion(client, branch)
	if baseVersion == 0 {
		return nil, fmt.Errorf("cannot reach game servers or no patches available for %s/%s (check firewall/network)", runtime.GOOS, runtime.GOARCH)
	}

	latestVersion := findLatestVersion(client, branch, baseVersion)

	versions := make([]int, 0, latestVersion-baseVersion+1)
	for v := baseVersion; v <= latestVersion; v++ {
		if versionExists(client, branch, v) {
			versions = append(versions, v)
		}
	}

	return versions, nil
}

func VerifyVersionExists(branch string, version int) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if versionExists(client, branch, version) {
		return nil
	}

	return fmt.Errorf("version %d not found", version)
}

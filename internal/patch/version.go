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

func FindLatestVersion(versionType string) (int, error) {
	key := fmt.Sprintf("%s-%s-%s", runtime.GOOS, runtime.GOARCH, versionType)

	if cached, ok := versionCache.get(key); ok {
		return cached.LatestVersion, cached.Error
	}

	result := checkVersion(versionType)
	versionCache.set(key, &result)

	return result.LatestVersion, result.Error
}

func ClearVersionCache() {
	versionCache.clear()
}

func checkVersion(versionType string) VersionCheckResult {
	client := createRobustClient()

	baseVersion := findBaseVersion(client, versionType)
	if baseVersion == 0 {
		return VersionCheckResult{
			Error: fmt.Errorf("cannot reach game servers or no patches available for %s/%s (check firewall/network)", runtime.GOOS, runtime.GOARCH),
		}
	}

	latestVersion := findLatestVersion(client, versionType, baseVersion)
	return VersionCheckResult{LatestVersion: latestVersion}
}

func createRobustClient() *http.Client {
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

func findBaseVersion(client *http.Client, versionType string) int {
	knownVersions := []int{25, 10, 5, 1}

	for _, v := range knownVersions {
		// Try each version with retry
		for attempt := 0; attempt < 2; attempt++ {
			if versionExists(client, versionType, v) {
				return v
			}
			if attempt < 1 {
				time.Sleep(time.Second)
			}
		}
	}
	return 0
}

func findLatestVersion(client *http.Client, versionType string, base int) int {
	if base <= 10 {
		return linearSearch(client, versionType, base, min(base+50, 200))
	}

	upper := exponentialSearch(client, versionType, base, 500)
	return binarySearch(client, versionType, base, upper)
}

func linearSearch(client *http.Client, versionType string, start, end int) int {
	latest := start
	for v := start + 1; v <= end; v++ {
		if versionExists(client, versionType, v) {
			latest = v
		} else {
			break
		}
	}
	return latest
}

func exponentialSearch(client *http.Client, versionType string, base, max int) int {
	current := base
	step := base

	for current < max {
		next := min(current+step, max)
		if versionExists(client, versionType, next) {
			current = next
			step *= 2
		} else {
			return current + step
		}
	}
	return max
}

func binarySearch(client *http.Client, versionType string, low, high int) int {
	latest := low

	for low < high {
		mid := (low + high + 1) / 2
		if versionExists(client, versionType, mid) {
			latest = mid
			low = mid
		} else {
			high = mid - 1
		}
	}
	return latest
}

func versionExists(client *http.Client, versionType string, version int) bool {
	url := fmt.Sprintf("https://game-patches.hytale.com/patches/%s/%s/%s/0/%d.pwr",
		runtime.GOOS, runtime.GOARCH, versionType, version)

	resp, err := client.Head(url)
	time.Sleep(200 * time.Millisecond)

	return err == nil && resp.StatusCode == http.StatusOK
}

func VerifyVersionExists(versionType string, version int) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if versionExists(client, versionType, version) {
		return nil
	}

	return fmt.Errorf("version %d not found", version)
}

package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Banner      string `json:"banner"`
	IP          string `json:"ip"`
}

type ServerWithUrls struct {
	Server
	LogoURL   string `json:"logo_url"`
	BannerURL string `json:"banner_url"`
}

type ServersService struct {
	apiBaseURL string
	cache      []ServerWithUrls
	cacheTime  time.Time
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
}

func NewServersService() *ServersService {
	return &ServersService{
		apiBaseURL: "https://api.hylauncher.fun",
		cacheTTL:   5 * time.Minute,
	}
}

func (s *ServersService) FetchServers() ([]ServerWithUrls, error) {
	// Check cache first
	s.cacheMu.RLock()
	if s.cache != nil && time.Since(s.cacheTime) < s.cacheTTL {
		cached := s.cache
		s.cacheMu.RUnlock()
		return cached, nil
	}
	s.cacheMu.RUnlock()

	// Fetch from API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(s.apiBaseURL + "/v1/servers")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("servers API returned status: %s", resp.Status)
	}

	var servers []Server
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		return nil, fmt.Errorf("failed to decode servers: %w", err)
	}

	result := make([]ServerWithUrls, len(servers))
	for i, server := range servers {
		result[i] = ServerWithUrls{
			Server:    server,
			LogoURL:   s.getUploadUrl(server.Logo),
			BannerURL: s.getUploadUrl(server.Banner),
		}
	}

	// Update cache
	s.cacheMu.Lock()
	s.cache = result
	s.cacheTime = time.Now()
	s.cacheMu.Unlock()

	return result, nil
}

func (s *ServersService) getUploadUrl(path string) string {
	if path == "" {
		return ""
	}
	if len(path) > 4 && path[:4] == "http" {
		return path
	}
	return s.apiBaseURL + path
}

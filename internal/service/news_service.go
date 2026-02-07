package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type NewsArticle struct {
	Title       string `json:"title"`
	DestURL     string `json:"dest_url"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

type NewsFeed struct {
	Articles []NewsArticle `json:"articles"`
}

type NewsService struct {
	feedURL string
}

func NewNewsService() *NewsService {
	return &NewsService{
		feedURL: "https://launcher.hytale.com/launcher-feed/release/feed.json",
	}
}

func (s *NewsService) FetchNews() ([]NewsArticle, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(s.feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("news feed returned status: %s", resp.Status)
	}

	var feed NewsFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to decode news feed: %w", err)
	}

	return feed.Articles, nil
}

func (s *NewsService) FetchLatestNews() (*NewsArticle, error) {
	articles, err := s.FetchNews()
	if err != nil {
		return nil, err
	}

	if len(articles) == 0 {
		return nil, fmt.Errorf("no news articles found in feed")
	}

	return &articles[0], nil
}

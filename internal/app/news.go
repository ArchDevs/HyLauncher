package app

import (
	"encoding/json"
	"net/http"
	"time"
)

type CoverImage struct {
	S3Key string `json:"s3Key"`
}

type BlogPost struct {
	ID          string     `json:"_id"`
	Title       string     `json:"title"`
	PublishedAt time.Time  `json:"publishedAt"`
	Slug        string     `json:"slug"`
	CoverImage  CoverImage `json:"coverImage"`
	BodyExcerpt string     `json:"bodyExcerpt"`
	Author      string     `json:"author"`
	Url         string     `json:"-"` // Constructed on frontend or here
	Image       string     `json:"-"` // Constructed on frontend or here
}

// GetNews fetches the latest blog posts from Hytale's official API
func (a *App) GetNews() ([]BlogPost, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://hytale.com/api/blog/post/published")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var posts []BlogPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}

	// Limit to latest 5 to save bandwidth/memory on frontend
	if len(posts) > 5 {
		posts = posts[:5]
	}

	return posts, nil
}

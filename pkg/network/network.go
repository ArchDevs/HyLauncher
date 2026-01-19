package network

import (
	"fmt"
	"net/http"
	"time"
)

func TestConnection(testURL string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head(testURL)
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error (HTTP %d)", resp.StatusCode)
	}

	return nil
}

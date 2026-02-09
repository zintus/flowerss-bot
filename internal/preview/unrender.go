package tgraph

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zintus/flowerss-bot/internal/config"
)

// FetchMarkdown fetches clean markdown content from the unrender service for the given URL.
// Returns an error if unrender is not configured or the request fails.
func FetchMarkdown(rawLink string) (string, error) {
	if config.UnrenderURL == "" || config.UnrenderToken == "" {
		return "", fmt.Errorf("unrender not configured")
	}

	url := config.UnrenderURL + "/" + rawLink

	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create unrender request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.UnrenderToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unrender request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unrender returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read unrender response: %w", err)
	}

	return string(body), nil
}

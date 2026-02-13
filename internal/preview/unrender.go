package tgraph

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/zintus/flowerss-bot/internal/config"
)

// FetchHTML fetches content from the unrender service and returns it as HTML
// suitable for Telegraph's CreatePageWithHTML. The unrender service returns
// markdown, which is converted to HTML via goldmark.
func FetchHTML(rawLink string) (string, error) {
	if config.UnrenderURL == "" || config.UnrenderToken == "" {
		return "", fmt.Errorf("unrender not configured")
	}

	requestURL := strings.TrimRight(config.UnrenderURL, "/") + "/" + url.PathEscape(rawLink)

	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("create unrender request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.UnrenderToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unrender request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unrender returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read unrender response: %w", err)
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(body, &buf); err != nil {
		return "", fmt.Errorf("markdown to html conversion: %w", err)
	}

	return buf.String(), nil
}

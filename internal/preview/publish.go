package tgraph

import (
	"html"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zintus/flowerss-bot/internal/log"
)

const (
	maxRetries       = 3
	baseBackoffSleep = 1 * time.Second
)

var floodWaitRegex = regexp.MustCompile(`FLOOD_WAIT_(\d+)`)

// parseFloodWaitSeconds extracts the wait seconds from a FLOOD_WAIT_X error.
// Returns 0 if the error is not a FLOOD_WAIT error.
func parseFloodWaitSeconds(err error) int {
	if err == nil {
		return 0
	}
	matches := floodWaitRegex.FindStringSubmatch(err.Error())
	if len(matches) >= 2 {
		if seconds, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			return seconds
		}
	}
	return 0
}

// isContentTooBig checks if the error indicates content is too large.
func isContentTooBig(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "CONTENT_TOO_BIG")
}

func PublishHtml(sourceTitle string, title string, rawLink string, htmlContent string) (string, error) {
	htmlContent = html.UnescapeString(htmlContent)

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		client := clientPool[rand.Intn(len(clientPool))]
		page, err := client.CreatePageWithHTML(
			title+" - "+sourceTitle, sourceTitle, rawLink, htmlContent, true,
		)
		if err == nil {
			zap.S().Infof("Created telegraph page url: %s", page.URL)
			return page.URL, nil
		}

		lastErr = err

		// Handle CONTENT_TOO_BIG - skip without retry
		if isContentTooBig(err) {
			log.Warnf("Create telegraph page failed: content too big, skipping (title: %s)", title)
			return "", nil
		}

		// Handle FLOOD_WAIT_X - wait the specified time before retry
		if waitSeconds := parseFloodWaitSeconds(err); waitSeconds > 0 {
			log.Warnf("Create telegraph page rate limited, waiting %d seconds before retry (attempt %d/%d)",
				waitSeconds, attempt+1, maxRetries)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		// For other errors, use exponential backoff
		if attempt < maxRetries-1 {
			backoff := baseBackoffSleep * time.Duration(1<<attempt)
			log.Warnf("Create telegraph page failed (attempt %d/%d), error: %s, retrying in %v",
				attempt+1, maxRetries, err, backoff)
			time.Sleep(backoff)
		}
	}

	log.Warnf("Create telegraph page failed after %d attempts, last error: %s", maxRetries, lastErr)
	return "", nil
}

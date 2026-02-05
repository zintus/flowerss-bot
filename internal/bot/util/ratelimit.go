package util

import (
	"errors"
	"time"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/log"
)

const (
	maxRetries       = 3
	baseBackoffSleep = 1 * time.Second
)

// SendWithRetry sends a message with automatic retry on rate limit errors.
// It handles Telegram's 429 (FloodError) by waiting the specified retry-after duration.
func SendWithRetry(ctx tb.Context, what interface{}, opts ...interface{}) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := ctx.Send(what, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error
		var floodErr tb.FloodError
		if errors.As(err, &floodErr) {
			waitSeconds := floodErr.RetryAfter
			if waitSeconds <= 0 {
				waitSeconds = 1 // Minimum wait of 1 second
			}
			log.Warnf("Telegram rate limited, waiting %d seconds before retry (attempt %d/%d)",
				waitSeconds, attempt+1, maxRetries)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		// For other errors, use exponential backoff only if retryable
		// Most Telegram errors are not retryable, so we break immediately
		break
	}

	return lastErr
}

// EditWithRetry edits a message with automatic retry on rate limit errors.
func EditWithRetry(ctx tb.Context, what interface{}, opts ...interface{}) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := ctx.Edit(what, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error
		var floodErr tb.FloodError
		if errors.As(err, &floodErr) {
			waitSeconds := floodErr.RetryAfter
			if waitSeconds <= 0 {
				waitSeconds = 1
			}
			log.Warnf("Telegram rate limited on edit, waiting %d seconds before retry (attempt %d/%d)",
				waitSeconds, attempt+1, maxRetries)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		break
	}

	return lastErr
}

// ReplyWithRetry replies to a message with automatic retry on rate limit errors.
func ReplyWithRetry(ctx tb.Context, what interface{}, opts ...interface{}) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := ctx.Reply(what, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error
		var floodErr tb.FloodError
		if errors.As(err, &floodErr) {
			waitSeconds := floodErr.RetryAfter
			if waitSeconds <= 0 {
				waitSeconds = 1
			}
			log.Warnf("Telegram rate limited on reply, waiting %d seconds before retry (attempt %d/%d)",
				waitSeconds, attempt+1, maxRetries)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		break
	}

	return lastErr
}

// BotSendWithRetry sends a message via Bot.Send with automatic retry on rate limit errors.
// This is useful for background tasks like BroadcastNews where we don't have a Context.
func BotSendWithRetry(bot *tb.Bot, to tb.Recipient, what interface{}, opts ...interface{}) error {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := bot.Send(to, what, opts...)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if it's a rate limit error
		var floodErr tb.FloodError
		if errors.As(err, &floodErr) {
			waitSeconds := floodErr.RetryAfter
			if waitSeconds <= 0 {
				waitSeconds = 1
			}
			log.Warnf("Telegram rate limited on bot send, waiting %d seconds before retry (attempt %d/%d)",
				waitSeconds, attempt+1, maxRetries)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}

		break
	}

	return lastErr
}

package util

import (
	tb "gopkg.in/telebot.v3"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

// GetLangCode extracts the language code from the context
// Falls back to DefaultLanguage if not found
func GetLangCode(ctx tb.Context) string {
	langCode := DefaultLanguage
	if langVal := ctx.Get(middleware.UserLanguageKey); langVal != nil {
		if val, ok := langVal.(string); ok && val != "" {
			langCode = val
		}
	}
	return langCode
}
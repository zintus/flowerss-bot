package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/middleware"
)

// DefaultLanguage is the default language code used when no user language is set
const DefaultLanguage = "en"

// getLangCode retrieves the language code for a user, defaulting to DefaultLanguage if not found
func getLangCode(ctx tb.Context) string {
	langCode := DefaultLanguage
	if langVal := ctx.Get(middleware.UserLanguageKey); langVal != nil {
		if val, ok := langVal.(string); ok && val != "" {
			langCode = val
		}
	}
	return langCode
}
package util

import (
	tb "gopkg.in/telebot.v3"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback
const UserLanguageKey = "user_lang" // Key for storing user language in context

// GetLangCode extracts the language code from the context
// Falls back to DefaultLanguage if not found
func GetLangCode(ctx tb.Context) string {
	langCode := DefaultLanguage
	if langVal := ctx.Get(UserLanguageKey); langVal != nil {
		if val, ok := langVal.(string); ok && val != "" {
			langCode = val
		}
	}
	return langCode
}
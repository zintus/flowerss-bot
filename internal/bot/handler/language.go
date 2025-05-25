package handler

import (
	"context" // Added for context.Background()
	"fmt"
	"strings" // For strings.Join and other manipulations

	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	// "github.com/zintus/flowerss-bot/internal/model" // Not strictly needed for this simplified version

	tb "gopkg.in/telebot.v3"
)

type Language struct {
	core *core.Core
}

func NewLanguageHandler(core *core.Core) *Language {
	return &Language{core: core}
}

func (l *Language) Command() string {
	return "/language"
}

func (l *Language) Description() string {
	return i18n.Localize(util.DefaultLanguage, "language_command_desc")
}

func (l *Language) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	args := ctx.Args()

	if len(args) == 0 {
		// Simplified for initial creation
		currentUser, _ := l.core.GetUser(context.Background(), ctx.Chat().ID) // Used context.Background()
		currentUserLangCode := langCode                                     // Default to context lang (which itself defaults to 'en' or user's preference)
		if currentUser != nil && currentUser.LanguageCode != "" {
			currentUserLangCode = currentUser.LanguageCode
		}

		currentLangDisplay := i18n.Localize(langCode, "language_current_language_format", currentUserLangCode)

		availableLangCodes := i18n.AvailableLanguages()
		// For the display "English (en)", we'd need "lang_en_name" in each locale.
		// For now, just showing codes as per simplification.
		// If language_display_format is "%s (%s)", and we want to show "English (en)"
		// we would need i18n.Localize(code, "lang_en_name") which means "lang_en_name" key in "en.json"
		// and i18n.Localize(code, "lang_zh_name") which means "lang_zh_name" key in "zh.json" etc.
		// This structure is not fully in place. So, we will list codes directly.
		// The "language_display_format" key might be used later if we enhance this.
		
		var langDisplays []string
		for _, code := range availableLangCodes {
			// Simple display: just the code.
			// To use language_display_format: e.g., fmt.Sprintf("%s (%s)", i18n.Localize(code, "lang_name"), code)
			// This implies a "lang_name" key in each language file (e.g. "lang_name": "English" in en.json)
			// For this subtask, keeping it simple and just listing codes.
			langDisplays = append(langDisplays, fmt.Sprintf("- %s", code))
		}


		availableLangsDisplay := i18n.Localize(langCode, "language_list_header") + "\n" + strings.Join(langDisplays, "\n")

		reply := currentLangDisplay + "\n\n" + availableLangsDisplay
		return ctx.Send(reply)
	}

	targetLangCode := strings.ToLower(args[0])
	isValid := false
	for _, code := range i18n.AvailableLanguages() {
		if targetLangCode == code {
			isValid = true
			break
		}
	}

	if !isValid {
		// For better UX, list available languages.
		availableLangCodes := i18n.AvailableLanguages()
		availableLangsList := strings.Join(availableLangCodes, ", ")
		errMsg := i18n.Localize(langCode, "language_set_fail_invalid_code_format", targetLangCode)
		// Appending available languages to the error message.
		// This could be part of the language_set_fail_invalid_code_format itself if we add another %s.
		// For now, simple concatenation.
		fullErrMsg := fmt.Sprintf("%s\n%s %s", errMsg, i18n.Localize(langCode, "language_list_header"), availableLangsList)
		return ctx.Reply(fullErrMsg)
	}

	err := l.core.SetUserLanguage(context.Background(), ctx.Chat().ID, targetLangCode) // Used context.Background()
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "err_system_error")) // Reusing existing key
	}

	// Send confirmation in the NEW language.
	// The success message "Language updated to %s." expects the language name.
	// As per simplification, we use the code.
	// If we had "lang_en_name": "English" in en.json, we could do:
	// langNameForSuccess := i18n.Localize(targetLangCode, "lang_"+targetLangCode+"_name")
	// For now, just targetLangCode
	return ctx.Reply(i18n.Localize(targetLangCode, "language_set_success_format", targetLangCode))
}

func (l *Language) Middlewares() []tb.MiddlewareFunc {
	return nil
}

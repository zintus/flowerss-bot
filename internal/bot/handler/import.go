package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

type Import struct {
}

func NewImport() *Import {
	return &Import{}
}

func (i *Import) Command() string {
	return "/import"
}

func getLangCode(ctx tb.Context) string {
	langCode := DefaultLanguage
	if langVal := ctx.Get(middleware.UserLanguageKey); langVal != nil {
		if val, ok := langVal.(string); ok && val != "" {
			langCode = val
		}
	}
	return langCode
}

func (i *Import) Description() string {
	return i18n.Localize(DefaultLanguage, "import_command_desc")
}

func (i *Import) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	reply := i18n.Localize(langCode, "import_handle_instruction")
	return ctx.Reply(reply)
}

func (i *Import) Middlewares() []tb.MiddlewareFunc {
	return nil
}

package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/config"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

type Version struct {
}

func NewVersion() *Version {
	return &Version{}
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

func (c *Version) Command() string {
	return "/version"
}

func (c *Version) Description() string {
	return i18n.Localize(DefaultLanguage, "version_command_desc")
}

func (c *Version) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	return ctx.Send(config.AppVersionInfo(langCode))
}

func (c *Version) Middlewares() []tb.MiddlewareFunc {
	return nil
}

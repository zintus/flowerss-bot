package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/config"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

type Version struct {
}

func NewVersion() *Version {
	return &Version{}
}

func (c *Version) Command() string {
	return "/version"
}

func (c *Version) Description() string {
	return i18n.Localize(util.DefaultLanguage, "version_command_desc")
}

func (c *Version) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	return ctx.Send(config.AppVersionInfo(langCode))
}

func (c *Version) Middlewares() []tb.MiddlewareFunc {
	return nil
}

package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

type Help struct {
}

func NewHelp() *Help {
	return &Help{}
}

func (h *Help) Command() string {
	return "/help"
}

func (h *Help) Description() string {
	return i18n.Localize(util.DefaultLanguage, "help_command_desc") // Using DefaultLanguage for command descriptions
}

func (h *Help) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	helpText := i18n.Localize(langCode, "help_message_text")
	return ctx.Send(helpText)
}

func (h *Help) Middlewares() []tb.MiddlewareFunc {
	return nil
}

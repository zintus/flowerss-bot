package handler

import (
	tb "gopkg.in/telebot.v3"

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
	return i18n.Localize("en", "help_command_desc") // Assuming "en" for descriptions
}

func (h *Help) Handle(ctx tb.Context) error {
	// TODO: Replace "en" with the actual user's language preference when available
	helpText := i18n.Localize("en", "help_message_text")
	return ctx.Send(helpText)
}

func (h *Help) Middlewares() []tb.MiddlewareFunc {
	return nil
}

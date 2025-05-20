package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/i18n"
)

type Ping struct {
	bot *tb.Bot
}

// NewPing new ping cmd handler
func NewPing(bot *tb.Bot) *Ping {
	return &Ping{bot: bot}
}

func (p *Ping) Command() string {
	return "/ping"
}

func (p *Ping) Description() string {
	return i18n.Localize("en", "ping_command_desc") // Assuming "en" for descriptions
}

func (p *Ping) Handle(ctx tb.Context) error {
	// TODO: Replace "en" with the actual user's language preference when available
	responseText := i18n.Localize("en", "ping_response_text")
	return ctx.Send(responseText)
}

func (p *Ping) Middlewares() []tb.MiddlewareFunc {
	return nil
}

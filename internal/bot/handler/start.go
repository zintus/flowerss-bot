package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
)

type Start struct {
}

func NewStart() *Start {
	return &Start{}
}

func (s *Start) Command() string {
	return "/start"
}

func (s *Start) Description() string {
	return i18n.Localize("en", "start_command_desc") // Assuming "en" for descriptions for now
}

func (s *Start) Handle(ctx tb.Context) error {
	log.Infof("/start id: %d", ctx.Chat().ID)
	// TODO: Replace "en" with the actual user's language preference when available
	welcomeMessage := i18n.Localize("en", "start_welcome_message")
	return ctx.Send(welcomeMessage)
}

func (s *Start) Middlewares() []tb.MiddlewareFunc {
	return nil
}

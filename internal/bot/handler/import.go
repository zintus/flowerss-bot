package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

// DefaultLanguage is defined in common.go

type Import struct {
}

func NewImport() *Import {
	return &Import{}
}

func (i *Import) Command() string {
	return "/import"
}

// getLangCode is defined in common.go

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

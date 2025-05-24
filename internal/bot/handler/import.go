package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

type Import struct {
}

func NewImport() *Import {
	return &Import{}
}

func (i *Import) Command() string {
	return "/import"
}

func (i *Import) Description() string {
	return i18n.Localize(util.DefaultLanguage, "import_command_desc")
}

func (i *Import) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	reply := i18n.Localize(langCode, "import_handle_instruction")
	return ctx.Reply(reply)
}

func (i *Import) Middlewares() []tb.MiddlewareFunc {
	return nil
}

package handler

import tb "gopkg.in/telebot.v3"

type Import struct {
}

func NewImport() *Import {
	return &Import{}
}

func (i *Import) Command() string {
	return "/import"
}

func (i *Import) Description() string {
	return "Import OPML file"
}

func (i *Import) Handle(ctx tb.Context) error {
	reply := "Please send the OPML file directly. If importing for a channel, include the channel ID when sending, e.g. @telegram"
	return ctx.Reply(reply)
}

func (i *Import) Middlewares() []tb.MiddlewareFunc {
	return nil
}

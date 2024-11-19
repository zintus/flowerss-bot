package handler

import (
	tb "gopkg.in/telebot.v3"
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
	return "Help"
}

func (h *Help) Handle(ctx tb.Context) error {
	message := `
	Commands:
	/sub Subscribe to RSS feed
	/unsub Unsubscribe from feed
	/list View current subscriptions
	/set Configure subscription settings
	/check Check current subscriptions
	/setfeedtag Set subscription tags
	/setinterval Set subscription refresh interval
	/activeall Activate all subscriptions
	/pauseall Pause all subscriptions
	/help Help
	/import Import OPML file
	/export Export OPML file
	/unsuball Unsubscribe from all feeds
	For detailed usage instructions visit: https://github.com/zintus/flowerss-bot
	`
	return ctx.Send(message)
}

func (h *Help) Middlewares() []tb.MiddlewareFunc {
	return nil
}

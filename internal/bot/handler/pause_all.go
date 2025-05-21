package handler

import (
	"context"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

type PauseAll struct {
	core *core.Core
}

func NewPauseAll(core *core.Core) *PauseAll {
	return &PauseAll{core: core}
}

func (p *PauseAll) Command() string {
	return "/pauseall"
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

func (p *PauseAll) Description() string {
	return i18n.Localize(DefaultLanguage, "pauseall_command_desc")
}

func (p *PauseAll) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	subscribeUserID := ctx.Message().Chat.ID
	var channelChat *tb.Chat
	v := ctx.Get(session.StoreKeyMentionChat.String())
	if v != nil {
		var ok bool
		channelChat, ok = v.(*tb.Chat)
		if ok && channelChat != nil {
			subscribeUserID = channelChat.ID
		}
	}

	source, err := p.core.GetUserSubscribedSources(context.Background(), subscribeUserID)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "err_system_error"))
	}

	for _, s := range source {
		err := p.core.DisableSourceUpdate(context.Background(), s.ID)
		if err != nil {
			return ctx.Reply(i18n.Localize(langCode, "pauseall_err_pause_failed"))
		}
	}

	var replyText string
	if channelChat != nil {
		replyText = i18n.Localize(langCode, "pauseall_success_channel", channelChat.Title, channelChat.Username)
	} else {
		replyText = i18n.Localize(langCode, "pauseall_success_user")
	}
	return ctx.Send(
		replyText, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             tb.ModeMarkdown,
		},
	)
}

func (p *PauseAll) Middlewares() []tb.MiddlewareFunc {
	return nil
}

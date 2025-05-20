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

type ActiveAll struct {
	core *core.Core
}

func NewActiveAll(core *core.Core) *ActiveAll {
	return &ActiveAll{core: core}
}

func (a *ActiveAll) Command() string {
	return "/activeall"
}

func (a *ActiveAll) Description() string {
	// Assuming "en" for command descriptions as they aren't user-specific yet in terms of language context
	return i18n.Localize(DefaultLanguage, "activeall_command_desc")
}

func (a *ActiveAll) Handle(ctx tb.Context) error {
	langCode := DefaultLanguage
	if langVal := ctx.Get(middleware.UserLanguageKey); langVal != nil {
		if val, ok := langVal.(string); ok && val != "" {
			langCode = val
		}
	}

	mentionChat, _ := session.GetMentionChatFromCtxStore(ctx)
	subscribeUserID := ctx.Chat().ID
	if mentionChat != nil {
		subscribeUserID = mentionChat.ID
	}

	source, err := a.core.GetUserSubscribedSources(context.Background(), subscribeUserID)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "err_system_error"))
	}

	for _, s := range source {
		err := a.core.EnableSourceUpdate(context.Background(), s.ID)
		if err != nil {
			return ctx.Reply(i18n.Localize(langCode, "activeall_err_activation_failed"))
		}
	}

	var reply string
	if mentionChat != nil {
		reply = i18n.Localize(langCode, "activeall_success_channel", mentionChat.Title, mentionChat.Username)
	} else {
		reply = i18n.Localize(langCode, "activeall_success_user")
	}

	return ctx.Reply(
		reply, &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             tb.ModeMarkdown,
		},
	)
}

func (a *ActiveAll) Middlewares() []tb.MiddlewareFunc {
	return nil
}

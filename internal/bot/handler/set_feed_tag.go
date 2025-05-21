package handler

import (
	"context"
	"strings"

	"github.com/spf13/cast"
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

type SetFeedTag struct {
	core *core.Core
}

func NewSetFeedTag(core *core.Core) *SetFeedTag {
	return &SetFeedTag{core: core}
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

func (s *SetFeedTag) Command() string {
	return "/setfeedtag"
}

func (s *SetFeedTag) Description() string {
	return i18n.Localize(DefaultLanguage, "setfeedtag_command_desc")
}

func (s *SetFeedTag) getMessageWithoutMention(ctx tb.Context) string {
	mention := message.MentionFromMessage(ctx.Message())
	if mention == "" {
		return ctx.Message().Payload
	}
	return strings.Replace(ctx.Message().Payload, mention, "", -1)
}

func (s *SetFeedTag) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	msg := s.getMessageWithoutMention(ctx)
	args := strings.Split(strings.TrimSpace(msg), " ")
	// Check if args[0] is empty, which means only command was sent or only mention
	if len(args) < 1 || args[0] == "" {
		return ctx.Reply(i18n.Localize(langCode, "setfeedtag_usage_hint"))
	}

	// 截短参数
	if len(args) > 4 {
		args = args[:4]
	}

	sourceID := cast.ToUint(args[0])
	mentionChat, _ := session.GetMentionChatFromCtxStore(ctx)
	subscribeUserID := ctx.Chat().ID
	if mentionChat != nil {
		subscribeUserID = mentionChat.ID
	}

	if err := s.core.SetSubscriptionTag(context.Background(), subscribeUserID, sourceID, args[1:]); err != nil {
		return ctx.Reply(i18n.Localize(langCode, "setfeedtag_err_set_failed"))
	}
	return ctx.Reply(i18n.Localize(langCode, "setfeedtag_success_set"))
}

func (s *SetFeedTag) Middlewares() []tb.MiddlewareFunc {
	return nil
}

package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
)

type SetUpdateInterval struct {
	core *core.Core
}

func NewSetUpdateInterval(core *core.Core) *SetUpdateInterval {
	return &SetUpdateInterval{core: core}
}

func (s *SetUpdateInterval) Command() string {
	return "/setinterval"
}

func (s *SetUpdateInterval) Description() string {
	return i18n.Localize(util.DefaultLanguage, "setinterval_command_desc")
}

func (s *SetUpdateInterval) getMessageWithoutMention(ctx tb.Context) string {
	mention := message.MentionFromMessage(ctx.Message())
	if mention == "" {
		return ctx.Message().Payload
	}
	return strings.ReplaceAll(ctx.Message().Payload, mention, "")
}

func (s *SetUpdateInterval) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	msg := s.getMessageWithoutMention(ctx)
	args := strings.Split(strings.TrimSpace(msg), " ")
	// Check if args[0] is empty, which means only command was sent or only mention
	if len(args) < 2 || args[0] == "" {
		return ctx.Reply(i18n.Localize(langCode, "setinterval_usage_hint"))
	}

	interval, err := strconv.Atoi(args[0])
	if interval <= 0 || err != nil {
		return ctx.Reply(i18n.Localize(langCode, "setinterval_err_invalid_interval"))
	}

	subscribeUserID := ctx.Message().Chat.ID
	mentionChat, _ := session.GetMentionChatFromCtxStore(ctx)
	if mentionChat != nil {
		subscribeUserID = mentionChat.ID
	}

	for _, id := range args[1:] {
		sourceID := cast.ToUint(id)
		if err := s.core.SetSubscriptionInterval(
			context.Background(), subscribeUserID, sourceID, interval,
		); err != nil {
			log.Errorf("SetSubscriptionInterval failed, %v", err)
			return ctx.Reply(i18n.Localize(langCode, "setinterval_err_set_failed"))
		}
	}
	return ctx.Reply(i18n.Localize(langCode, "setinterval_success_set"))
}

func (s *SetUpdateInterval) Middlewares() []tb.MiddlewareFunc {
	return nil
}

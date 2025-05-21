package handler

import (
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

const (
	SetSubscriptionTagButtonUnique = "set_set_sub_tag_btn"
	DefaultLanguage                = "en" // Define DefaultLanguage for fallback
)

type SetSubscriptionTagButton struct {
	bot *tb.Bot
}

func NewSetSubscriptionTagButton(bot *tb.Bot) *SetSubscriptionTagButton {
	return &SetSubscriptionTagButton{bot: bot}
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

func (b *SetSubscriptionTagButton) CallbackUnique() string {
	return "\f" + SetSubscriptionTagButtonUnique
}

func (b *SetSubscriptionTagButton) Description() string {
	return ""
}

func (b *SetSubscriptionTagButton) feedSetAuth(c *tb.Callback, attachData *session.Attachment) bool {
	subscriberID := attachData.GetUserId()
	if subscriberID != c.Sender.ID {
		channelChat, err := b.bot.ChatByID(subscriberID)
		if err != nil {
			return false
		}

		if !chat.IsChatAdmin(b.bot, channelChat, c.Sender.ID) {
			return false
		}
	}
	return true
}

func (b *SetSubscriptionTagButton) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	c := ctx.Callback()
	attachData, err := session.UnmarshalAttachment(ctx.Callback().Data)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "err_system_error"))
	}

	if !b.feedSetAuth(c, attachData) {
		// Using ctx.Send as per analysis in task description for permission errors
		return ctx.Send(i18n.Localize(langCode, "err_permission_denied"))
	}
	sourceID := uint(attachData.GetSourceId())
	msg := i18n.Localize(langCode, "setsubtagbtn_usage_hint_format", sourceID, sourceID)
	return ctx.Edit(msg, &tb.SendOptions{ParseMode: tb.ModeMarkdown})
}

func (b *SetSubscriptionTagButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

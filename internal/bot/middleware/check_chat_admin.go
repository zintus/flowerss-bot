package middleware

import (
	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/i18n"

	tb "gopkg.in/telebot.v3"
)

func IsChatAdmin() tb.MiddlewareFunc {
	return func(next tb.HandlerFunc) tb.HandlerFunc {
		return func(c tb.Context) error {
			langCode := util.GetLangCode(c)
			if !chat.IsChatAdmin(c.Bot(), c.Chat(), c.Sender().ID) {
				return c.Reply(i18n.Localize(langCode, "middleware_err_not_chat_admin"))
			}

			v := c.Get(session.StoreKeyMentionChat.String())
			if v != nil {
				mentionChat, ok := v.(*tb.Chat)
				if !ok {
					return c.Reply(i18n.Localize(langCode, "err_internal_error"))
				}
				if !chat.IsChatAdmin(c.Bot(), mentionChat, c.Sender().ID) {
					return c.Reply(i18n.Localize(langCode, "middleware_err_not_chat_admin"))
				}
			}
			return next(c)
		}
	}
}

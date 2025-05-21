package handler

import (
	"context"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

// DefaultLanguage is defined in common.go

type RemoveAllSubscription struct {
}

func NewRemoveAllSubscription() *RemoveAllSubscription {
	return &RemoveAllSubscription{}
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

func (r RemoveAllSubscription) Command() string {
	return "/unsuball"
}

func (r RemoveAllSubscription) Description() string {
	return i18n.Localize(DefaultLanguage, "unsuball_command_desc")
}

func (r RemoveAllSubscription) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	reply := i18n.Localize(langCode, "unsuball_confirm_message")
	var confirmKeys [][]tb.InlineButton
	confirmKeys = append(
		confirmKeys, []tb.InlineButton{
			{
				Unique: UnSubAllButtonUnique,
				Text:   i18n.Localize(langCode, "btn_confirm"),
			},
			{
				Unique: CancelUnSubAllButtonUnique,
				Text:   i18n.Localize(langCode, "btn_cancel"),
			},
		},
	)
	return ctx.Reply(reply, &tb.ReplyMarkup{InlineKeyboard: confirmKeys})
}

func (r RemoveAllSubscription) Middlewares() []tb.MiddlewareFunc {
	return nil
}

const (
	UnSubAllButtonUnique       = "unsub_all_confirm_btn"
	CancelUnSubAllButtonUnique = "unsub_all_cancel_btn"
)

type RemoveAllSubscriptionButton struct {
	core *core.Core
}

func NewRemoveAllSubscriptionButton(core *core.Core) *RemoveAllSubscriptionButton {
	return &RemoveAllSubscriptionButton{core: core}
}

func (r *RemoveAllSubscriptionButton) CallbackUnique() string {
	return "\f" + UnSubAllButtonUnique
}

func (r *RemoveAllSubscriptionButton) Description() string {
	return ""
}

func (r *RemoveAllSubscriptionButton) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	err := r.core.UnsubscribeAllSource(context.Background(), ctx.Sender().ID)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "unsuball_err_unsubscribe_failed"))
	}
	return ctx.Edit(i18n.Localize(langCode, "unsuball_success_unsubscribed"))
}

func (r *RemoveAllSubscriptionButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

type CancelRemoveAllSubscriptionButton struct {
}

func NewCancelRemoveAllSubscriptionButton() *CancelRemoveAllSubscriptionButton {
	return &CancelRemoveAllSubscriptionButton{}
}

func (r *CancelRemoveAllSubscriptionButton) CallbackUnique() string {
	return "\f" + CancelUnSubAllButtonUnique
}

func (r *CancelRemoveAllSubscriptionButton) Description() string {
	return ""
}

func (r *CancelRemoveAllSubscriptionButton) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	return ctx.Edit(i18n.Localize(langCode, "unsuball_info_operation_cancelled"))
}

func (r *CancelRemoveAllSubscriptionButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

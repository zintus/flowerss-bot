package handler

import (
	"bytes"
	"context"
	"text/template"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/config"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/model"
)

// TelegraphSwitchButtonUnique is defined in common.go

type TelegraphSwitchButton struct {
	bot  *tb.Bot
	core *core.Core
}

func NewTelegraphSwitchButton(bot *tb.Bot, core *core.Core) *TelegraphSwitchButton {
	return &TelegraphSwitchButton{bot: bot, core: core}
}

func (b *TelegraphSwitchButton) CallbackUnique() string {
	return "\f" + TelegraphSwitchButtonUnique
}

func (b *TelegraphSwitchButton) Description() string {
	return ""
}

func (b *TelegraphSwitchButton) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	c := ctx.Callback()
	if c == nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_callback_nil")})
	}

	attachData, err := session.UnmarshalAttachment(ctx.Callback().Data)
	if err != nil {
		// Using notify_switch_err_generic as per task instruction for unmarshal error in ctx.Respond context
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}
	subscriberID := attachData.GetUserId()
	if subscriberID != c.Sender.ID {
		channelChat, err := b.bot.ChatByID(subscriberID)
		if err != nil {
			return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
		}
		if !chat.IsChatAdmin(b.bot, channelChat, c.Sender.ID) {
			return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
		}
	}

	sourceID := uint(attachData.GetSourceId())
	source, _ := b.core.GetSource(context.Background(), sourceID) // Error ignored in original

	err = b.core.ToggleSubscriptionTelegraph(context.Background(), subscriberID, sourceID)
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	sub, err := b.core.GetSubscription(context.Background(), subscriberID, sourceID)
	if sub == nil || err != nil { // Original code checks sub == nil OR err != nil
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	// feedSettingTmpl is defined in common.go
	// Use common getTemplateFuncMap and feedSettingTmpl
	t := template.New("setting template").Funcs(getTemplateFuncMap(langCode))
	_, err = t.Parse(feedSettingTmpl) // feedSettingTmpl is now from common.go
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	text := new(bytes.Buffer)
	err = t.Execute(text, map[string]interface{}{"source": source, "sub": sub, "Count": config.ErrorThreshold})
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	_ = ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "subswitch_success_updated")})
	
	// Use genFeedSetBtn from common.go
	return ctx.Edit(
		text.String(),
		&tb.SendOptions{ParseMode: tb.ModeHTML},
		&tb.ReplyMarkup{InlineKeyboard: genFeedSetBtn(c, sub, source, langCode)},
	)
}

// genFeedSetBtnFromSet is removed, using genFeedSetBtn from common.go

func (b *TelegraphSwitchButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

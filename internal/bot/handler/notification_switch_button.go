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

const (
	NotificationSwitchButtonUnique = "set_toggle_notice_btn"
)

// Use the same template as in other handlers
const feedSettingTmpl = `
{{ .L "set_tmpl_header_settings" }}
{{ .L "set_tmpl_label_id" }} {{ .source.ID }}
{{ .L "set_tmpl_label_title" }} {{ .source.Title }}
{{ .L "set_tmpl_label_link" }} {{ .source.Link }}
{{ .L "set_tmpl_label_updates" }} {{if ge .source.ErrorCount .Count }}{{ .L "set_tmpl_status_paused" }}{{else}}{{ .L "set_tmpl_status_active" }}{{end}}
{{ .L "set_tmpl_label_interval" }} {{ .sub.Interval }} {{ .L "set_tmpl_unit_minutes" }}
{{ .L "set_tmpl_label_notifications" }} {{if eq .sub.EnableNotification 0}}{{ .L "set_tmpl_status_off" }}{{else}}{{ .L "set_tmpl_status_on" }}{{end}}
{{ .L "set_tmpl_label_telegraph" }} {{if eq .sub.EnableTelegraph 0}}{{ .L "set_tmpl_status_off" }}{{else}}{{ .L "set_tmpl_status_on" }}{{end}}
{{ .L "set_tmpl_label_tags" }} {{if .sub.Tag}}{{ .sub.Tag }}{{else}}{{ .L "set_tmpl_status_none" }}{{end}}
`

type NotificationSwitchButton struct {
	bot  *tb.Bot
	core *core.Core
}

func NewNotificationSwitchButton(bot *tb.Bot, core *core.Core) *NotificationSwitchButton {
	return &NotificationSwitchButton{bot: bot, core: core}
}

// Using the shared utility function instead of local implementation

func (b *NotificationSwitchButton) CallbackUnique() string {
	return "\f" + NotificationSwitchButtonUnique
}

func (b *NotificationSwitchButton) Description() string {
	return ""
}

func (b *NotificationSwitchButton) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	c := ctx.Callback()
	if c == nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_callback_nil")})
	}

	attachData, err := session.UnmarshalAttachment(ctx.Callback().Data)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "err_system_error"))
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
	source, err := b.core.GetSource(context.Background(), sourceID)
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	err = b.core.ToggleSubscriptionNotice(context.Background(), subscriberID, sourceID)
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	sub, err := b.core.GetSubscription(context.Background(), subscriberID, sourceID)
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	funcMap := template.FuncMap{
		"L": func(key string, args ...interface{}) string {
			return i18n.Localize(langCode, key, args...)
		},
	}
	t := template.New("setting template").Funcs(funcMap)
	_, err = t.Parse(feedSettingTmpl)
	if err != nil {
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	text := new(bytes.Buffer)
	err = t.Execute(text, map[string]interface{}{"source": source, "sub": sub, "Count": config.ErrorThreshold})
	if err != nil {
		// Log error, return generic message
		return ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_err_generic")})
	}

	_ = ctx.Respond(&tb.CallbackResponse{Text: i18n.Localize(langCode, "notify_switch_success_updated")})
	return ctx.Edit(
		text.String(),
		&tb.SendOptions{ParseMode: tb.ModeHTML},
		&tb.ReplyMarkup{InlineKeyboard: genFeedSetBtnFromSet(c, sub, source, langCode)},
	)
}

func (b *NotificationSwitchButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

// Wrapper for the genFeedSetBtn function from set.go
func genFeedSetBtnFromSet(c *tb.Callback, sub *model.Subscribe, source *model.Source, langCode string) [][]tb.InlineButton {
	// Create buttons with the same constants as in set.go
	setSubTagKey := tb.InlineButton{
		Unique: "set_set_sub_tag_btn", // SetSubscriptionTagButtonUnique
		Text:   i18n.Localize(langCode, "set_btn_tag_settings"),
		Data:   c.Data,
	}

	var notificationTextKey string
	if sub.EnableNotification == 1 {
		notificationTextKey = "set_btn_disable_notifications"
	} else {
		notificationTextKey = "set_btn_enable_notifications"
	}
	toggleNoticeKey := tb.InlineButton{
		Unique: NotificationSwitchButtonUnique,
		Text:   i18n.Localize(langCode, notificationTextKey),
		Data:   c.Data,
	}

	var telegraphTextKey string
	if sub.EnableTelegraph == 1 {
		telegraphTextKey = "set_btn_disable_telegraph"
	} else {
		telegraphTextKey = "set_btn_enable_telegraph"
	}
	toggleTelegraphKey := tb.InlineButton{
		Unique: "set_toggle_telegraph_btn", // TelegraphSwitchButtonUnique
		Text:   i18n.Localize(langCode, telegraphTextKey),
		Data:   c.Data,
	}

	var updatesTextKey string
	if source.ErrorCount >= config.ErrorThreshold {
		updatesTextKey = "set_btn_resume_updates"
	} else {
		updatesTextKey = "set_btn_pause_updates"
	}
	toggleEnabledKey := tb.InlineButton{
		Unique: "set_toggle_update_btn", // SubscriptionSwitchButtonUnique
		Text:   i18n.Localize(langCode, updatesTextKey),
		Data:   c.Data,
	}

	feedSettingKeys := [][]tb.InlineButton{
		{
			toggleEnabledKey,
			toggleNoticeKey,
		},
		{
			toggleTelegraphKey,
			setSubTagKey,
		},
	}
	return feedSettingKeys
}
package handler

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/config"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/model"
)

// DefaultLanguage is defined in common.go

type Set struct {
	bot  *tb.Bot
	core *core.Core
}

func NewSet(bot *tb.Bot, core *core.Core) *Set {
	return &Set{
		bot:  bot,
		core: core,
	}
}

// getLangCode is defined in common.go

func (s *Set) Command() string {
	return "/set"
}

func (s *Set) Description() string {
	return i18n.Localize(DefaultLanguage, "set_command_desc")
}

func (s *Set) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	mentionChat, _ := session.GetMentionChatFromCtxStore(ctx)
	ownerID := ctx.Message().Chat.ID
	if mentionChat != nil {
		ownerID = mentionChat.ID
	}

	sources, err := s.core.GetUserSubscribedSources(context.Background(), ownerID)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "set_err_get_subs_failed"))
	}
	if len(sources) <= 0 {
		return ctx.Reply(i18n.Localize(langCode, "set_info_no_subs"))
	}

	setFeedItemBtns := [][]tb.InlineButton{}
	for _, source := range sources {
		attachData := &session.Attachment{
			UserId:   ctx.Chat().ID, // Note: this might be different from ownerID if mentionChat is used
			SourceId: uint32(source.ID),
		}

		data := session.Marshal(attachData)
		setFeedItemBtns = append(
			setFeedItemBtns, []tb.InlineButton{
				{
					Unique: SetFeedItemButtonUnique,
					Text:   fmt.Sprintf("[%d] %s", source.ID, source.Title), // Button text can remain as is
					Data:   data,
				},
			},
		)
	}

	return ctx.Reply(
		i18n.Localize(langCode, "set_info_select_feed_to_configure"), &tb.ReplyMarkup{
			InlineKeyboard: setFeedItemBtns,
		},
	)
}

func (s *Set) Middlewares() []tb.MiddlewareFunc {
	return nil
}

const (
	SetFeedItemButtonUnique = "set_feed_item_btn"
	feedSettingTmpl         = `
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
)

type SetFeedItemButton struct {
	bot  *tb.Bot
	core *core.Core
}

func NewSetFeedItemButton(bot *tb.Bot, core *core.Core) *SetFeedItemButton {
	return &SetFeedItemButton{bot: bot, core: core}
}

func (r *SetFeedItemButton) CallbackUnique() string {
	return "\f" + SetFeedItemButtonUnique
}

func (r *SetFeedItemButton) Description() string {
	return ""
}

func (r *SetFeedItemButton) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	attachData, err := session.UnmarshalAttachment(ctx.Callback().Data)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "set_err_button_settings_error"))
	}

	subscriberID := attachData.GetUserId()
	if subscriberID != ctx.Callback().Sender.ID {
		channelChat, err := r.bot.ChatByUsername(fmt.Sprintf("%d", subscriberID))
		if err != nil {
			return ctx.Edit(i18n.Localize(langCode, "set_err_get_sub_info_failed_button"))
		}

		if !chat.IsChatAdmin(r.bot, channelChat, ctx.Callback().Sender.ID) {
			return ctx.Edit(i18n.Localize(langCode, "set_err_get_sub_info_failed_button")) // Or a more specific permission error
		}
	}

	sourceID := uint(attachData.GetSourceId())
	source, err := r.core.GetSource(context.Background(), sourceID)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "set_err_source_not_found"))
	}

	sub, err := r.core.GetSubscription(context.Background(), subscriberID, source.ID)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "set_err_user_not_subscribed"))
	}

	funcMap := template.FuncMap{
		"L": func(key string, args ...interface{}) string {
			return i18n.Localize(langCode, key, args...)
		},
	}
	t := template.New("setting template").Funcs(funcMap)
	_, err = t.Parse(feedSettingTmpl)
	if err != nil {
		// Log error, return generic message
		return ctx.Edit(i18n.Localize(langCode, "set_err_button_settings_error"))
	}

	text := new(bytes.Buffer)
	err = t.Execute(text, map[string]interface{}{"source": source, "sub": sub, "Count": config.ErrorThreshold})
	if err != nil {
		// Log error, return generic message
		return ctx.Edit(i18n.Localize(langCode, "set_err_button_settings_error"))
	}

	return ctx.Edit(
		text.String(),
		&tb.SendOptions{ParseMode: tb.ModeHTML},
		&tb.ReplyMarkup{InlineKeyboard: genFeedSetBtn(ctx.Callback(), sub, source, langCode)},
	)
}

func genFeedSetBtn(
	c *tb.Callback, sub *model.Subscribe, source *model.Source, langCode string,
) [][]tb.InlineButton {
	setSubTagKey := tb.InlineButton{
		Unique: SetSubscriptionTagButtonUnique,
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
		Unique: TelegraphSwitchButtonUnique,
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
		Unique: SubscriptionSwitchButtonUnique,
		Text:   i18n.Localize(langCode, updatesTextKey),
		Data:   c.Data,
	}

	feedSettingKeys := [][]tb.InlineButton{
		{ // Row 1
			toggleEnabledKey,
			toggleNoticeKey,
		},
		{ // Row 2
			toggleTelegraphKey,
			setSubTagKey,
		},
	}
	return feedSettingKeys
}

func (r *SetFeedItemButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

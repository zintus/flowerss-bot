// Package handler provides telegram bot command handlers
package handler

import (
	"text/template"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/config"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/model"
)

// Common constants for button uniques
const (
	SubscriptionSwitchButtonUnique = "set_toggle_update_btn"
	SetSubscriptionTagButtonUnique = "set_set_sub_tag_btn"
	NotificationSwitchButtonUnique = "set_toggle_notice_btn"
	TelegraphSwitchButtonUnique    = "set_toggle_telegraph_btn"
	SetFeedItemButtonUnique        = "set_feed_item_btn" // From set.go
)

// Common template for feed settings
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

// Common function to generate feed setting buttons
// Ensure all handlers that use this function and feedSettingTmpl correctly initialize
// the template with getTemplateFuncMap(langCode).
func genFeedSetBtn(
	c *tb.Callback, sub *model.Subscribe, source *model.Source, langCode string,
) [][]tb.InlineButton {
	setSubTagKey := tb.InlineButton{
		Unique: SetSubscriptionTagButtonUnique, // Uses common constant
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
		Unique: NotificationSwitchButtonUnique, // Uses common constant
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
		Unique: TelegraphSwitchButtonUnique, // Uses common constant
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
		Unique: SubscriptionSwitchButtonUnique, // Uses common constant
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

// getTemplateFuncMap provides the template.FuncMap for rendering the feedSettingTmpl.
// Each handler should use this to ensure "L" function is available for localization.
func getTemplateFuncMap(langCode string) template.FuncMap {
	return template.FuncMap{
		"L": func(key string, args ...interface{}) string {
			return i18n.Localize(langCode, key, args...)
		},
	}
}
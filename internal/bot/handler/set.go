package handler

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/config"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
)

// Use util.DefaultLanguage instead of local declaration

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

// Use util.GetLangCode instead of local implementation

func (s *Set) Command() string {
	return "/set"
}

func (s *Set) Description() string {
	return i18n.Localize(util.DefaultLanguage, "set_command_desc")
}

func (s *Set) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
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

// SetFeedItemButtonUnique is defined in common.go
// feedSettingTmpl is defined in common.go

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
	langCode := util.GetLangCode(ctx)
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

	// Use common getTemplateFuncMap and feedSettingTmpl
	t := template.New("setting template").Funcs(getTemplateFuncMap(langCode))
	_, err = t.Parse(feedSettingTmpl) // feedSettingTmpl is now from common.go
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
		// Use genFeedSetBtn from common.go
		&tb.ReplyMarkup{InlineKeyboard: genFeedSetBtn(ctx.Callback(), sub, source, langCode)},
	)
}

// genFeedSetBtn is defined in common.go

func (r *SetFeedItemButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

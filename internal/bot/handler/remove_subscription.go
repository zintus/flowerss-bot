package handler

import (
	"context"
	"fmt"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
)

// DefaultLanguage is defined in common.go

type RemoveSubscription struct {
	bot  *tb.Bot
	core *core.Core
}

func NewRemoveSubscription(bot *tb.Bot, core *core.Core) *RemoveSubscription {
	return &RemoveSubscription{
		bot:  bot,
		core: core,
	}
}

// getLangCode is defined in common.go

func (s *RemoveSubscription) Command() string {
	return "/unsub"
}

func (s *RemoveSubscription) Description() string {
	return i18n.Localize(DefaultLanguage, "unsub_command_desc")
}

func (s *RemoveSubscription) removeForChannel(ctx tb.Context, channelName string) error {
	langCode := getLangCode(ctx)
	sourceURL := message.URLFromMessage(ctx.Message())
	if sourceURL == "" {
		return ctx.Send(i18n.Localize(langCode, "unsub_hint_channel_usage"))
	}

	channelChat, err := s.bot.ChatByUsername(channelName)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "err_get_channel_info_failed"))
	}

	if !chat.IsChatAdmin(s.bot, channelChat, ctx.Sender().ID) {
		return ctx.Reply(i18n.Localize(langCode, "err_admin_only_operation"))
	}

	source, err := s.core.GetSourceByURL(context.Background(), sourceURL)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "unsub_err_get_sub_info_failed"))
	}

	log.Infof("%d for [%d]%s unsubscribe %s", ctx.Chat().ID, source.ID, source.Title, source.Link)
	if err := s.core.Unsubscribe(context.Background(), channelChat.ID, source.ID); err != nil {
		log.Errorf(
			"%d for [%d]%s unsubscribe %s failed, %v",
			ctx.Chat().ID, source.ID, source.Title, source.Link, err,
		)
		return ctx.Reply(i18n.Localize(langCode, "unsub_err_unsubscribe_failed"))
	}
	return ctx.Send(
		i18n.Localize(langCode, "unsub_success_channel_format", channelChat.Title, channelChat.Username, source.Title, source.Link),
		&tb.SendOptions{DisableWebPagePreview: true, ParseMode: tb.ModeMarkdown},
	)
}

func (s *RemoveSubscription) removeForChat(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	sourceURL := message.URLFromMessage(ctx.Message())
	if sourceURL == "" {
		sources, err := s.core.GetUserSubscribedSources(context.Background(), ctx.Chat().ID)
		if err != nil {
			return ctx.Reply(i18n.Localize(langCode, "unsub_err_get_sub_list_failed"))
		}

		if len(sources) == 0 {
			return ctx.Reply(i18n.Localize(langCode, "unsub_info_no_subscriptions"))
		}

		var unsubFeedItemButtons [][]tb.InlineButton
		for _, source := range sources {
			attachData := &session.Attachment{
				UserId:   ctx.Chat().ID,
				SourceId: uint32(source.ID),
			}

			data := session.Marshal(attachData)
			unsubFeedItemButtons = append(
				unsubFeedItemButtons, []tb.InlineButton{
					{
						Unique: RemoveSubscriptionItemButtonUnique,
						Text:   fmt.Sprintf("[%d] %s", source.ID, source.Title), // Button text can remain as is, or be localized if needed
						Data:   data,
					},
				},
			)
		}
		return ctx.Reply(i18n.Localize(langCode, "unsub_info_select_feed_to_unsub"), &tb.ReplyMarkup{InlineKeyboard: unsubFeedItemButtons})
	}

	if !chat.IsChatAdmin(s.bot, ctx.Chat(), ctx.Sender().ID) {
		return ctx.Reply(i18n.Localize(langCode, "err_admin_only_operation"))
	}

	source, err := s.core.GetSourceByURL(context.Background(), sourceURL)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "unsub_err_not_subscribed_feed"))
	}

	log.Infof("%d unsubscribe [%d]%s %s", ctx.Chat().ID, source.ID, source.Title, source.Link)
	if err := s.core.Unsubscribe(context.Background(), ctx.Chat().ID, source.ID); err != nil {
		log.Errorf(
			"%d for [%d]%s unsubscribe %s failed, %v",
			ctx.Chat().ID, source.ID, source.Title, source.Link, err,
		)
		return ctx.Reply(i18n.Localize(langCode, "unsub_err_unsubscribe_failed"))
	}
	return ctx.Send(
		i18n.Localize(langCode, "unsub_success_user_format", source.Title, source.Link),
		&tb.SendOptions{DisableWebPagePreview: true, ParseMode: tb.ModeMarkdown},
	)
}

func (s *RemoveSubscription) Handle(ctx tb.Context) error {
	mention := message.MentionFromMessage(ctx.Message())
	if mention != "" {
		return s.removeForChannel(ctx, mention)
	}
	return s.removeForChat(ctx)
}

func (s *RemoveSubscription) Middlewares() []tb.MiddlewareFunc {
	return nil
}

const (
	RemoveSubscriptionItemButtonUnique = "unsub_feed_item_btn"
)

type RemoveSubscriptionItemButton struct {
	core *core.Core
}

func NewRemoveSubscriptionItemButton(core *core.Core) *RemoveSubscriptionItemButton {
	return &RemoveSubscriptionItemButton{core: core}
}

func (r *RemoveSubscriptionItemButton) CallbackUnique() string {
	return "\f" + RemoveSubscriptionItemButtonUnique
}

func (r *RemoveSubscriptionItemButton) Description() string {
	return ""
}

func (r *RemoveSubscriptionItemButton) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx) // Add langCode retrieval
	if ctx.Callback() == nil {
		return ctx.Edit(i18n.Localize(langCode, "err_internal_error"))
	}

	attachData, err := session.UnmarshalAttachment(ctx.Callback().Data)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "unsub_err_button_action_failed"))
	}

	userID := attachData.GetUserId()
	sourceID := uint(attachData.GetSourceId())
	source, err := r.core.GetSource(context.Background(), sourceID)
	if err != nil {
		return ctx.Edit(i18n.Localize(langCode, "unsub_err_button_action_failed"))
	}

	if err := r.core.Unsubscribe(context.Background(), userID, sourceID); err != nil {
		log.Errorf("unsubscribe data %s failed, %v", ctx.Callback().Data, err)
		return ctx.Edit(i18n.Localize(langCode, "unsub_err_button_action_failed"))
	}

	rtnMsg := i18n.Localize(langCode, "unsub_success_button_format", sourceID, source.Link, source.Title)
	return ctx.Edit(rtnMsg, &tb.SendOptions{ParseMode: tb.ModeHTML})
}

func (r *RemoveSubscriptionItemButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

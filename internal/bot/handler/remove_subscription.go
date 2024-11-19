package handler

import (
	"context"
	"fmt"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/log"
)

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

func (s *RemoveSubscription) Command() string {
	return "/unsub"
}

func (s *RemoveSubscription) Description() string {
	return "Unsubscribe from RSS feed"
}

func (s *RemoveSubscription) removeForChannel(ctx tb.Context, channelName string) error {
	sourceURL := message.URLFromMessage(ctx.Message())
	if sourceURL == "" {
		return ctx.Send("To unsubscribe a channel, use the command: '/unsub @ChannelID URL'")
	}

	channelChat, err := s.bot.ChatByUsername(channelName)
	if err != nil {
		return ctx.Reply("Error getting channel information")
	}

	if !chat.IsChatAdmin(s.bot, channelChat, ctx.Sender().ID) {
		return ctx.Reply("Only channel administrators can perform this operation")
	}

	source, err := s.core.GetSourceByURL(context.Background(), sourceURL)
	if err != nil {
		return ctx.Reply("Error getting subscription information")
	}

	log.Infof("%d for [%d]%s unsubscribe %s", ctx.Chat().ID, source.ID, source.Title, source.Link)
	if err := s.core.Unsubscribe(context.Background(), channelChat.ID, source.ID); err != nil {
		log.Errorf(
			"%d for [%d]%s unsubscribe %s failed, %v",
			ctx.Chat().ID, source.ID, source.Title, source.Link, err,
		)
		return ctx.Reply("Unsubscribe failed")
	}
	return ctx.Send(
		fmt.Sprintf(
			"Channel [%s](https://t.me/%s) successfully unsubscribed from [%s](%s)",
			channelChat.Title, channelChat.Username, source.Title, source.Link,
		),
		&tb.SendOptions{DisableWebPagePreview: true, ParseMode: tb.ModeMarkdown},
	)
}

func (s *RemoveSubscription) removeForChat(ctx tb.Context) error {
	sourceURL := message.URLFromMessage(ctx.Message())
	if sourceURL == "" {
		sources, err := s.core.GetUserSubscribedSources(context.Background(), ctx.Chat().ID)
		if err != nil {
			return ctx.Reply("Failed to get subscription list")
		}

		if len(sources) == 0 {
			return ctx.Reply("No subscriptions")
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
						Text:   fmt.Sprintf("[%d] %s", source.ID, source.Title),
						Data:   data,
					},
				},
			)
		}
		return ctx.Reply("Please select the feed you want to unsubscribe from", &tb.ReplyMarkup{InlineKeyboard: unsubFeedItemButtons})
	}

	if !chat.IsChatAdmin(s.bot, ctx.Chat(), ctx.Sender().ID) {
		return ctx.Reply("Only administrators can perform this operation")
	}

	source, err := s.core.GetSourceByURL(context.Background(), sourceURL)
	if err != nil {
		return ctx.Reply("Not subscribed to this RSS feed")
	}

	log.Infof("%d unsubscribe [%d]%s %s", ctx.Chat().ID, source.ID, source.Title, source.Link)
	if err := s.core.Unsubscribe(context.Background(), ctx.Chat().ID, source.ID); err != nil {
		log.Errorf(
			"%d for [%d]%s unsubscribe %s failed, %v",
			ctx.Chat().ID, source.ID, source.Title, source.Link, err,
		)
		return ctx.Reply("Unsubscribe failed")
	}
	return ctx.Send(
		fmt.Sprintf("[%s](%s) successfully unsubscribed!", source.Title, source.Link),
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
	if ctx.Callback() == nil {
		return ctx.Edit("Internal error!")
	}

	attachData, err := session.UnmarshalAttachment(ctx.Callback().Data)
	if err != nil {
		return ctx.Edit("Unsubscribe error!")
	}

	userID := attachData.GetUserId()
	sourceID := uint(attachData.GetSourceId())
	source, err := r.core.GetSource(context.Background(), sourceID)
	if err != nil {
		return ctx.Edit("Unsubscribe error!")
	}

	if err := r.core.Unsubscribe(context.Background(), userID, sourceID); err != nil {
		log.Errorf("unsubscribe data %s failed, %v", ctx.Callback().Data, err)
		return ctx.Edit("Unsubscribe error!")
	}

	rtnMsg := fmt.Sprintf("[%d] <a href=\"%s\">%s</a> successfully unsubscribed", sourceID, source.Link, source.Title)
	return ctx.Edit(rtnMsg, &tb.SendOptions{ParseMode: tb.ModeHTML})
}

func (r *RemoveSubscriptionItemButton) Middlewares() []tb.MiddlewareFunc {
	return nil
}

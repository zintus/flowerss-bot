package handler

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/chat"
	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
	"github.com/zintus/flowerss-bot/internal/model"
)

const (
	MaxSubsSizePerPage = 50
)

type ListSubscription struct {
	core *core.Core
}

func NewListSubscription(core *core.Core) *ListSubscription {
	return &ListSubscription{core: core}
}

// Using the shared utility function instead of local implementation

func (l *ListSubscription) Command() string {
	return "/list"
}

func (l *ListSubscription) Description() string {
	return i18n.Localize(util.DefaultLanguage, "listsub_command_desc")
}

func (l *ListSubscription) listChatSubscription(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	// private chat or group
	if ctx.Chat().Type != tb.ChatPrivate && !chat.IsChatAdmin(ctx.Bot(), ctx.Chat(), ctx.Sender().ID) {
		return ctx.Send(i18n.Localize(langCode, "err_permission_denied"))
	}

	stdCtx := context.Background()
	sources, err := l.core.GetUserSubscribedSources(stdCtx, ctx.Chat().ID)
	if err != nil {
		log.Errorf("GetUserSubscribedSources failed, %v", err)
		return ctx.Send(i18n.Localize(langCode, "listsub_err_get_subs_failed"))
	}

	return l.replaySubscribedSources(ctx, sources, langCode)
}

func (l *ListSubscription) listChannelSubscription(ctx tb.Context, channelName string) error {
	langCode := util.GetLangCode(ctx)
	channelChat, err := ctx.Bot().ChatByUsername(channelName)
	if err != nil {
		return ctx.Send(i18n.Localize(langCode, "err_get_channel_info_failed"))
	}

	if !chat.IsChatAdmin(ctx.Bot(), channelChat, ctx.Sender().ID) {
		return ctx.Send(i18n.Localize(langCode, "err_not_channel_admin_action"))
	}

	stdCtx := context.Background()
	sources, err := l.core.GetUserSubscribedSources(stdCtx, channelChat.ID)
	if err != nil {
		log.Errorf("GetUserSubscribedSources failed, %v", err)
		return ctx.Send(i18n.Localize(langCode, "listsub_err_get_subs_failed"))
	}
	return l.replaySubscribedSources(ctx, sources, langCode)
}

func (l *ListSubscription) Handle(ctx tb.Context) error {
	mention := message.MentionFromMessage(ctx.Message())
	if mention != "" {
		return l.listChannelSubscription(ctx, mention)
	}
	return l.listChatSubscription(ctx)
}

func (l *ListSubscription) Middlewares() []tb.MiddlewareFunc {
	return nil
}

func (l *ListSubscription) replaySubscribedSources(ctx tb.Context, sources []*model.Source, langCode string) error {
	if len(sources) == 0 {
		return ctx.Send(i18n.Localize(langCode, "listsub_info_sub_list_empty"))
	}

	sort.Slice(sources, func(i, j int) bool {
		a := sources[i].LastPublishedAt
		b := sources[j].LastPublishedAt
		if a == nil && b == nil {
			return false // Maintain original order if both are nil
		}
		if a == nil {
			return false // Nil dates come last
		}
		if b == nil {
			return true // Nil dates come last
		}
		return a.After(*b)
	})
	var msg strings.Builder
	msg.WriteString(i18n.Localize(langCode, "listsub_list_header_format", len(sources)))
	count := 0
	for i := range sources {
		publishedDate := "N/A"
		if sources[i].LastPublishedAt != nil {
			publishedDate = sources[i].LastPublishedAt.Format("2006-01-02 15:04:05")
		}
		msg.WriteString(fmt.Sprintf("[[%d]] [%s](%s) - %s\n", sources[i].ID, sources[i].Title, sources[i].Link, publishedDate))
		count++
		if count == MaxSubsSizePerPage {
			if err := ctx.Send(msg.String(), &tb.SendOptions{DisableWebPagePreview: true, ParseMode: tb.ModeMarkdown}); err != nil {
				log.Errorf("failed to send subscription list page: %v", err)
			}
			count = 0
			msg.Reset()
		}
	}

	if count != 0 {
		if err := ctx.Send(msg.String(), &tb.SendOptions{DisableWebPagePreview: true, ParseMode: tb.ModeMarkdown}); err != nil {
			log.Errorf("failed to send final subscription list page: %v", err)
		}
	}
	return nil
}

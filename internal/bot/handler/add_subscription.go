package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

var (
	ErrGetChannelInfoFailedForPerms = errors.New("failed to get channel info for permissions")
)

type AddSubscription struct {
	core *core.Core
}

func NewAddSubscription(core *core.Core) *AddSubscription {
	return &AddSubscription{
		core: core,
	}
}

func (a *AddSubscription) Command() string {
	return "/sub"
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

func (a *AddSubscription) Description() string {
	return i18n.Localize(DefaultLanguage, "addsub_command_desc")
}

func (a *AddSubscription) addSubscriptionForChat(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	sourceURL := message.URLFromMessage(ctx.Message())
	if sourceURL == "" {
		hint := i18n.Localize(langCode, "addsub_hint_no_url_chat", a.Command())
		return ctx.Send(hint, &tb.SendOptions{ReplyTo: ctx.Message()})
	}

	source, err := a.core.CreateSource(context.Background(), sourceURL)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_create_source_failed_format", err.Error()))
	}

	log.Infof("%d subscribe [%d]%s %s", ctx.Chat().ID, source.ID, source.Title, source.Link)
	if err := a.core.AddSubscription(context.Background(), ctx.Chat().ID, source.ID); err != nil {
		if errors.Is(err, core.ErrSubscriptionExist) {
			return ctx.Reply(i18n.Localize(langCode, "addsub_err_already_subscribed"))
		}
		log.Errorf("add subscription user %d source %d failed %v", ctx.Chat().ID, source.ID, err)
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_generic_subscribe_failed"))
	}

	return ctx.Reply(
		i18n.Localize(langCode, "addsub_success_subscribed_format", source.ID, source.Title, source.Link),
		&tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             tb.ModeMarkdown,
		},
	)
}

func (a *AddSubscription) hasChannelPrivilege(bot *tb.Bot, channelChat *tb.Chat, opUserID int64, botID int64) (
	bool, error,
) {
	adminList, err := bot.AdminsOf(channelChat)
	if err != nil {
		zap.S().Error(err)
		return false, ErrGetChannelInfoFailedForPerms
	}

	senderIsAdmin := false
	botIsAdmin := false
	for _, admin := range adminList {
		if opUserID == admin.User.ID {
			senderIsAdmin = true
		}
		if botID == admin.User.ID {
			botIsAdmin = true
		}
	}

	return botIsAdmin && senderIsAdmin, nil
}

func (a *AddSubscription) addSubscriptionForChannel(ctx tb.Context, channelName string) error {
	langCode := getLangCode(ctx)
	sourceURL := message.URLFromMessage(ctx.Message())
	if sourceURL == "" {
		return ctx.Send(i18n.Localize(langCode, "addsub_hint_no_url_channel"))
	}

	bot := ctx.Bot()
	channelChat, err := bot.ChatByUsername(channelName)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "err_get_channel_info_failed"))
	}
	if channelChat.Type != tb.ChatChannel {
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_not_channel_admin"))
	}

	hasPrivilege, errPriv := a.hasChannelPrivilege(bot, channelChat, ctx.Sender().ID, bot.Me.ID)
	if errPriv != nil {
		if errors.Is(errPriv, ErrGetChannelInfoFailedForPerms) {
			return ctx.Reply(i18n.Localize(langCode, "err_get_channel_info_failed"))
		}
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_generic_subscribe_failed"))
	}
	if !hasPrivilege {
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_not_channel_admin"))
	}

	source, err := a.core.CreateSource(context.Background(), sourceURL)
	if err != nil {
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_create_source_failed_format", err.Error()))
	}

	log.Infof("%d subscribe [%d]%s %s", channelChat.ID, source.ID, source.Title, source.Link)
	if err := a.core.AddSubscription(context.Background(), channelChat.ID, source.ID); err != nil {
		if errors.Is(err, core.ErrSubscriptionExist) {
			return ctx.Reply(i18n.Localize(langCode, "addsub_err_already_subscribed"))
		}
		log.Errorf("add subscription user %d source %d failed %v", channelChat.ID, source.ID, err)
		return ctx.Reply(i18n.Localize(langCode, "addsub_err_generic_subscribe_failed"))
	}

	return ctx.Reply(
		i18n.Localize(langCode, "addsub_success_subscribed_format", source.ID, source.Title, source.Link),
		&tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             tb.ModeMarkdown,
		},
	)
}

func (a *AddSubscription) Handle(ctx tb.Context) error {
	mention := message.MentionFromMessage(ctx.Message())
	if mention != "" {
		// has mention, add subscription for channel
		return a.addSubscriptionForChannel(ctx, mention)
	}
	return a.addSubscriptionForChat(ctx)
}

func (a *AddSubscription) Middlewares() []tb.MiddlewareFunc {
	return nil
}

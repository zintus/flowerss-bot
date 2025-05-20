package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/message"
	"github.com/zintus/flowerss-bot/internal/bot/middleware"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
	"github.com/zintus/flowerss-bot/internal/model"
	"github.com/zintus/flowerss-bot/internal/opml"
)

const DefaultLanguage = "en" // Define DefaultLanguage for fallback

var (
	ErrExportGetChannelInfo      = errors.New("export: unable to get channel information")
	ErrExportGetChannelAdminInfo = errors.New("export: unable to get channel admin information")
	ErrExportChannelAdminOnly    = errors.New("export: only channel admins can perform this operation")
	ErrExportGetSourceInfo       = errors.New("export: failed to get source information")
)

type Export struct {
	core *core.Core
}

func NewExport(core *core.Core) *Export {
	return &Export{core: core}
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

func (e *Export) Description() string {
	return i18n.Localize(DefaultLanguage, "export_command_desc")
}

func (e *Export) Command() string {
	return "/export"
}

func (e *Export) getChannelSources(bot *tb.Bot, opUserID int64, channelName string) ([]*model.Source, error) {
	channelChat, err := bot.ChatByUsername(channelName)
	if err != nil {
		return nil, ErrExportGetChannelInfo
	}

	adminList, err := bot.AdminsOf(channelChat)
	if err != nil {
		return nil, ErrExportGetChannelAdminInfo
	}

	senderIsAdmin := false
	for _, admin := range adminList {
		if opUserID == admin.User.ID {
			senderIsAdmin = true
			break
		}
	}

	if !senderIsAdmin {
		return nil, ErrExportChannelAdminOnly
	}

	sources, err := e.core.GetUserSubscribedSources(context.Background(), channelChat.ID)
	if err != nil {
		zap.S().Error(err) // Keep original logging
		return nil, ErrExportGetSourceInfo
	}
	return sources, nil
}

func (e *Export) Handle(ctx tb.Context) error {
	langCode := getLangCode(ctx)
	mention := message.MentionFromMessage(ctx.Message())
	var sources []*model.Source
	var err error

	if mention == "" {
		sources, err = e.core.GetUserSubscribedSources(context.Background(), ctx.Chat().ID)
		if err != nil {
			log.Error(err)
			return ctx.Send(i18n.Localize(langCode, "export_err_generic_export_failed"))
		}
	} else {
		sources, err = e.getChannelSources(ctx.Bot(), ctx.Chat().ID, mention)
		if err != nil {
			log.Error(err) // Keep the log
			var errKey string
			switch {
			case errors.Is(err, ErrExportGetChannelInfo):
				errKey = "export_err_get_channel_info"
			case errors.Is(err, ErrExportGetChannelAdminInfo):
				errKey = "export_err_get_channel_admin_info"
			case errors.Is(err, ErrExportChannelAdminOnly):
				errKey = "export_err_channel_admin_only"
			case errors.Is(err, ErrExportGetSourceInfo):
				errKey = "export_err_get_source_info"
			default:
				errKey = "export_err_generic_export_failed"
			}
			return ctx.Send(i18n.Localize(langCode, errKey))
		}
	}

	if len(sources) == 0 {
		return ctx.Send(i18n.Localize(langCode, "export_info_sub_list_empty"))
	}

	opmlStr, err := opml.ToOPML(sources)
	if err != nil {
		log.Error(err) // Keep original logging for OPML generation error
		return ctx.Send(i18n.Localize(langCode, "export_err_generic_export_failed"))
	}
	opmlFile := &tb.Document{File: tb.FromReader(strings.NewReader(opmlStr))}
	opmlFile.FileName = fmt.Sprintf("subscriptions_%d.opml", time.Now().Unix())
	if err := ctx.Send(opmlFile); err != nil {
		log.Errorf("send OPML file failed, err:%v", err)
		return ctx.Send(i18n.Localize(langCode, "export_err_generic_export_failed"))
	}
	return nil
}

func (e *Export) Middlewares() []tb.MiddlewareFunc {
	return nil
}

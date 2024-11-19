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
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/log"
	"github.com/zintus/flowerss-bot/internal/model"
	"github.com/zintus/flowerss-bot/internal/opml"
)

type Export struct {
	core *core.Core
}

func NewExport(core *core.Core) *Export {
	return &Export{core: core}
}

func (e *Export) Description() string {
	return "Export OPML"
}

func (e *Export) Command() string {
	return "/export"
}

func (e *Export) getChannelSources(bot *tb.Bot, opUserID int64, channelName string) ([]*model.Source, error) {
	// 导出channel订阅
	channelChat, err := bot.ChatByUsername(channelName)
	if err != nil {
		return nil, errors.New("Unable to get channel information")
	}

	adminList, err := bot.AdminsOf(channelChat)
	if err != nil {
		return nil, errors.New("Unable to get channel administrator information")
	}

	senderIsAdmin := false
	for _, admin := range adminList {
		if opUserID == admin.User.ID {
			senderIsAdmin = true
			break
		}
	}

	if !senderIsAdmin {
		return nil, errors.New("Only channel administrators can perform this operation")
	}

	sources, err := e.core.GetUserSubscribedSources(context.Background(), channelChat.ID)
	if err != nil {
		zap.S().Error(err)
		return nil, errors.New("Failed to get subscription source information")
	}
	return sources, nil
}

func (e *Export) Handle(ctx tb.Context) error {
	mention := message.MentionFromMessage(ctx.Message())
	var sources []*model.Source
	if mention == "" {
		var err error
		sources, err = e.core.GetUserSubscribedSources(context.Background(), ctx.Chat().ID)
		if err != nil {
			log.Error(err)
			return ctx.Send("Export failed")
		}
	} else {
		var err error
		sources, err = e.getChannelSources(ctx.Bot(), ctx.Chat().ID, mention)
		if err != nil {
			log.Error(err)
			return ctx.Send("导出失败")
		}
	}

	if len(sources) == 0 {
		return ctx.Send("Subscription list is empty")
	}

	opmlStr, err := opml.ToOPML(sources)
	if err != nil {
		return ctx.Send("Export failed")
	}
	opmlFile := &tb.Document{File: tb.FromReader(strings.NewReader(opmlStr))}
	opmlFile.FileName = fmt.Sprintf("subscriptions_%d.opml", time.Now().Unix())
	if err := ctx.Send(opmlFile); err != nil {
		log.Errorf("send OPML file failed, err:%v", err)
		return ctx.Send("导出失败")
	}
	return nil
}

func (e *Export) Middlewares() []tb.MiddlewareFunc {
	return nil
}

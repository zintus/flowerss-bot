package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/log"
	"github.com/zintus/flowerss-bot/internal/opml"

	tb "gopkg.in/telebot.v3"
)

var (
	ErrNotOPMLFile   = errors.New("ondocument: not an opml file")
	ErrGetFileFailed = errors.New("ondocument: failed to retrieve file")
)

type OnDocument struct {
	bot  *tb.Bot
	core *core.Core
}

func NewOnDocument(bot *tb.Bot, core *core.Core) *OnDocument {
	return &OnDocument{
		bot:  bot,
		core: core,
	}
}

func (o *OnDocument) Command() string {
	return tb.OnDocument
}

func (o *OnDocument) Description() string {
	return "" // This is tb.OnDocument, not a user-visible description string
}

func (o *OnDocument) getOPML(ctx tb.Context) (*opml.OPML, error) {
	if !strings.HasSuffix(ctx.Message().Document.FileName, ".opml") {
		return nil, ErrNotOPMLFile
	}

	fileRead, err := o.bot.File(&ctx.Message().Document.File)
	if err != nil {
		return nil, ErrGetFileFailed // Wrapped error or direct return? For now, direct.
	}

	opmlFile, err := opml.ReadOPML(fileRead)
	if err != nil {
		log.Errorf("parser opml failed, %v", err) // Keep original logging
		return nil, ErrGetFileFailed              // Assuming parsing error means file retrieval/read issue at a high level
	}
	return opmlFile, nil
}

func (o *OnDocument) Handle(ctx tb.Context) error {
	langCode := util.GetLangCode(ctx)
	opmlFile, err := o.getOPML(ctx)
	if err != nil {
		if errors.Is(err, ErrNotOPMLFile) {
			return ctx.Reply(i18n.Localize(langCode, "ondoc_err_not_opml"))
		} else if errors.Is(err, ErrGetFileFailed) {
			return ctx.Reply(i18n.Localize(langCode, "ondoc_err_get_file_failed"))
		}
		// For any other error type from getOPML, if any, pass through original message.
		return ctx.Reply(err.Error())
	}

	userID := ctx.Chat().ID
	v := ctx.Get(session.StoreKeyMentionChat.String())
	if mentionChat, ok := v.(*tb.Chat); ok && mentionChat != nil {
		userID = mentionChat.ID
	}

	outlines, _ := opmlFile.GetFlattenOutlines()
	var failImportList []opml.Outline
	var successImportList []opml.Outline
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, outline := range outlines {
		outline := outline
		wg.Add(1)
		go func() {
			defer wg.Done()
			source, err := o.core.CreateSource(context.Background(), outline.XMLURL)
			if err != nil {
				mu.Lock()
				failImportList = append(failImportList, outline)
				mu.Unlock()
				return
			}

			err = o.core.AddSubscription(context.Background(), userID, source.ID)
			if err != nil {
				mu.Lock()
				if errors.Is(err, core.ErrSubscriptionExist) {
					successImportList = append(successImportList, outline)
				} else {
					failImportList = append(failImportList, outline)
				}
				mu.Unlock()
				return
			}

			log.Infof("%d subscribe [%d]%s %s", ctx.Chat().ID, source.ID, source.Title, source.Link)
			mu.Lock()
			successImportList = append(successImportList, outline)
			mu.Unlock()
		}()
	}
	wg.Wait()

	var msg strings.Builder
	msg.WriteString(i18n.Localize(langCode, "ondoc_import_summary_format", len(successImportList), len(failImportList)))
	if len(successImportList) != 0 {
		msg.WriteString(i18n.Localize(langCode, "ondoc_import_success_header"))
		for i, line := range successImportList {
			if line.Text != "" {
				msg.WriteString(
					fmt.Sprintf("[%d] <a href=\"%s\">%s</a>\n", i+1, line.XMLURL, line.Text),
				)
			} else {
				msg.WriteString(fmt.Sprintf("[%d] %s\n", i+1, line.XMLURL))
			}
		}
		msg.WriteString("\n")
	}

	if len(failImportList) != 0 {
		msg.WriteString(i18n.Localize(langCode, "ondoc_import_failure_header"))
		for i, line := range failImportList {
			if line.Text != "" {
				msg.WriteString(fmt.Sprintf("[%d] <a href=\"%s\">%s</a>\n", i+1, line.XMLURL, line.Text))
			} else {
				msg.WriteString(fmt.Sprintf("[%d] %s\n", i+1, line.XMLURL))
			}
		}
	}

	return ctx.Reply(
		msg.String(), &tb.SendOptions{
			DisableWebPagePreview: true,
			ParseMode:             tb.ModeHTML,
		},
	)
}

func (o *OnDocument) Middlewares() []tb.MiddlewareFunc {
	return nil
}

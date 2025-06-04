package handler

import (
	"context"
	"testing"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/session"
	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/i18n"
	"github.com/zintus/flowerss-bot/internal/model"
	"github.com/zintus/flowerss-bot/internal/storage"
)

type editCaptureCtx struct {
	tb.Context
	last string
}

func (e *editCaptureCtx) Edit(what interface{}, opts ...interface{}) error {
	if s, ok := what.(string); ok {
		e.last = s
	}
	return nil
}

// --- mock storages ---
type mockSourceStorage struct {
	getSourceFunc func(ctx context.Context, id uint) (*model.Source, error)
}

func (m *mockSourceStorage) Init(ctx context.Context) error { return nil }

func (m *mockSourceStorage) AddSource(ctx context.Context, source *model.Source) error {
	panic("not implemented")
}
func (m *mockSourceStorage) GetSource(ctx context.Context, id uint) (*model.Source, error) {
	if m.getSourceFunc != nil {
		return m.getSourceFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockSourceStorage) GetSources(ctx context.Context) ([]*model.Source, error) {
	panic("not implemented")
}
func (m *mockSourceStorage) GetSourceByURL(ctx context.Context, url string) (*model.Source, error) {
	panic("not implemented")
}
func (m *mockSourceStorage) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockSourceStorage) UpsertSource(ctx context.Context, sourceID uint, newSource *model.Source) error {
	panic("not implemented")
}

type mockSubscriptionStorage struct {
	existFunc  func(ctx context.Context, userID int64, sourceID uint) (bool, error)
	deleteFunc func(ctx context.Context, userID int64, sourceID uint) (int64, error)
	countFunc  func(ctx context.Context, sourceID uint) (int64, error)
}

func (m *mockSubscriptionStorage) Init(ctx context.Context) error { return nil }

func (m *mockSubscriptionStorage) AddSubscription(ctx context.Context, sub *model.Subscribe) error {
	panic("not implemented")
}
func (m *mockSubscriptionStorage) SubscriptionExist(ctx context.Context, userID int64, sourceID uint) (bool, error) {
	if m.existFunc != nil {
		return m.existFunc(ctx, userID, sourceID)
	}
	return false, nil
}
func (m *mockSubscriptionStorage) GetSubscription(ctx context.Context, userID int64, sourceID uint) (*model.Subscribe, error) {
	panic("not implemented")
}
func (m *mockSubscriptionStorage) GetSubscriptionsByUserID(ctx context.Context, userID int64, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error) {
	panic("not implemented")
}
func (m *mockSubscriptionStorage) GetSubscriptionsBySourceID(ctx context.Context, sourceID uint, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error) {
	panic("not implemented")
}
func (m *mockSubscriptionStorage) CountSubscriptions(ctx context.Context) (int64, error) {
	panic("not implemented")
}
func (m *mockSubscriptionStorage) DeleteSubscription(ctx context.Context, userID int64, sourceID uint) (int64, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userID, sourceID)
	}
	return 1, nil
}
func (m *mockSubscriptionStorage) CountSourceSubscriptions(ctx context.Context, sourceID uint) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, sourceID)
	}
	return 1, nil
}
func (m *mockSubscriptionStorage) UpdateSubscription(ctx context.Context, userID int64, sourceID uint, newSubscription *model.Subscribe) error {
	panic("not implemented")
}
func (m *mockSubscriptionStorage) UpsertSubscription(ctx context.Context, userID int64, sourceID uint, newSubscription *model.Subscribe) error {
	panic("not implemented")
}

// dummy content storage
type mockContentStorage struct{}

func (m *mockContentStorage) Init(ctx context.Context) error { return nil }

func (m *mockContentStorage) AddContent(ctx context.Context, content *model.Content) error {
	return nil
}
func (m *mockContentStorage) DeleteSourceContents(ctx context.Context, sourceID uint) (int64, error) {
	return 0, nil
}
func (m *mockContentStorage) HashIDExist(ctx context.Context, hashID string) (bool, error) {
	return false, nil
}

// dummy user storage
type mockUserStorage struct{}

func (m *mockUserStorage) Init(ctx context.Context) error                         { return nil }
func (m *mockUserStorage) CreateUser(ctx context.Context, user *model.User) error { return nil }
func (m *mockUserStorage) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return nil, nil
}
func (m *mockUserStorage) SetUserLanguage(ctx context.Context, userID int64, langCode string) error {
	return nil
}

func TestRemoveSubscriptionItemButton_Handle(t *testing.T) {
	i18n.ResetTranslationsForTest()
	if err := i18n.LoadTranslations("../../locales"); err != nil {
		t.Fatalf("load translations: %v", err)
	}

	userID := int64(123)
	sourceID := uint(1)

	mockSrc := &mockSourceStorage{
		getSourceFunc: func(ctx context.Context, id uint) (*model.Source, error) {
			return &model.Source{ID: id, Link: "http://example.com", Title: "Example"}, nil
		},
	}
	unsubCalled := false
	mockSub := &mockSubscriptionStorage{
		existFunc: func(ctx context.Context, u int64, s uint) (bool, error) { return true, nil },
		deleteFunc: func(ctx context.Context, u int64, s uint) (int64, error) {
			if u != userID || s != sourceID {
				t.Errorf("unexpected ids")
			}
			unsubCalled = true
			return 1, nil
		},
		countFunc: func(ctx context.Context, s uint) (int64, error) { return 1, nil },
	}
	c := core.NewCore(&mockUserStorage{}, &mockContentStorage{}, mockSrc, mockSub, nil, nil)

	bot, err := tb.NewBot(tb.Settings{Token: "TEST", Offline: true})
	if err != nil {
		t.Fatalf("new bot: %v", err)
	}
	att := &session.Attachment{UserId: userID, SourceId: uint32(sourceID)}
	data := session.Marshal(att)

	msg := &tb.Message{ID: 1, Chat: &tb.Chat{ID: userID}}
	cb := &tb.Callback{ID: "cb", Data: data, Message: msg, Sender: &tb.User{ID: userID}}
	update := tb.Update{ID: 1, Callback: cb}
	baseCtx := bot.NewContext(update)

	ctx := &editCaptureCtx{Context: baseCtx}
	ctx.Set(util.UserLanguageKey, "en")

	h := NewRemoveSubscriptionItemButton(c)
	if err := h.Handle(ctx); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if !unsubCalled {
		t.Errorf("expected unsubscribe called")
	}
	want := i18n.Localize("en", "unsub_success_button_format", sourceID, "http://example.com", "Example")
	if ctx.last != want {
		t.Errorf("expected '%s', got '%s'", want, ctx.last)
	}
}

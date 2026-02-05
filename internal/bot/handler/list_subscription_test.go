package handler

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/model"
	"github.com/zintus/flowerss-bot/internal/storage"
)

type mockListSubCtx struct {
	tb.Context
	sentMessages []string
}

func (m *mockListSubCtx) Send(what interface{}, opts ...interface{}) error {
	if s, ok := what.(string); ok {
		m.sentMessages = append(m.sentMessages, s)
	}
	return nil
}

type mockListSubSubscriptionStorage struct {
	getUserSubsFunc func(ctx context.Context, userID int64, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error)
}

func (m *mockListSubSubscriptionStorage) Init(ctx context.Context) error { return nil }

func (m *mockListSubSubscriptionStorage) AddSubscription(ctx context.Context, sub *model.Subscribe) error {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) SubscriptionExist(ctx context.Context, userID int64, sourceID uint) (bool, error) {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) GetSubscription(ctx context.Context, userID int64, sourceID uint) (*model.Subscribe, error) {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) GetSubscriptionsByUserID(ctx context.Context, userID int64, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error) {
	if m.getUserSubsFunc != nil {
		return m.getUserSubsFunc(ctx, userID, opts)
	}
	return &storage.GetSubscriptionsResult{}, nil
}

func (m *mockListSubSubscriptionStorage) GetSubscriptionsBySourceID(ctx context.Context, sourceID uint, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error) {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) CountSubscriptions(ctx context.Context) (int64, error) {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) DeleteSubscription(ctx context.Context, userID int64, sourceID uint) (int64, error) {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) CountSourceSubscriptions(ctx context.Context, sourceID uint) (int64, error) {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) UpdateSubscription(ctx context.Context, userID int64, sourceID uint, newSubscription *model.Subscribe) error {
	panic("not implemented")
}

func (m *mockListSubSubscriptionStorage) UpsertSubscription(ctx context.Context, userID int64, sourceID uint, newSubscription *model.Subscribe) error {
	panic("not implemented")
}

type mockListSubSourceStorage struct {
	getSourceFunc func(ctx context.Context, id uint) (*model.Source, error)
}

func (m *mockListSubSourceStorage) Init(ctx context.Context) error { return nil }

func (m *mockListSubSourceStorage) AddSource(ctx context.Context, source *model.Source) error {
	panic("not implemented")
}

func (m *mockListSubSourceStorage) GetSource(ctx context.Context, id uint) (*model.Source, error) {
	if m.getSourceFunc != nil {
		return m.getSourceFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockListSubSourceStorage) GetSources(ctx context.Context) ([]*model.Source, error) {
	panic("not implemented")
}

func (m *mockListSubSourceStorage) GetSourceByURL(ctx context.Context, url string) (*model.Source, error) {
	panic("not implemented")
}

func (m *mockListSubSourceStorage) Delete(ctx context.Context, id uint) error { return nil }

func (m *mockListSubSourceStorage) UpsertSource(ctx context.Context, sourceID uint, newSource *model.Source) error {
	panic("not implemented")
}

func TestListSubscription_replaySubscribedSources_Sorted(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	sources := map[uint]*model.Source{
		1: {ID: 1, Title: "Feed C", Link: "http://example.com/c", LastContentAt: &yesterday},
		2: {ID: 2, Title: "Feed A", Link: "http://example.com/a", LastContentAt: &now},
		3: {ID: 3, Title: "Feed B", Link: "http://example.com/b", LastContentAt: &twoDaysAgo},
		4: {ID: 4, Title: "Feed D", Link: "http://example.com/d", LastContentAt: nil}, // nil date
	}

	subscriptions := []*model.Subscribe{
		{UserID: 123, SourceID: 1},
		{UserID: 123, SourceID: 2},
		{UserID: 123, SourceID: 3},
		{UserID: 123, SourceID: 4},
	}

	sourceStorage := &mockListSubSourceStorage{
		getSourceFunc: func(ctx context.Context, id uint) (*model.Source, error) {
			if source, ok := sources[id]; ok {
				return source, nil
			}
			return nil, storage.ErrRecordNotFound
		},
	}

	subStorage := &mockListSubSubscriptionStorage{
		getUserSubsFunc: func(ctx context.Context, userID int64, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error) {
			return &storage.GetSubscriptionsResult{
				Subscriptions: subscriptions,
			}, nil
		},
	}

	coreInstance := core.NewCore(nil, nil, sourceStorage, subStorage, nil, nil)
	h := NewListSubscription(coreInstance)

	ctx := &mockListSubCtx{
		Context: &fakeContext{
			chatID: 123,
			data:   map[string]interface{}{util.UserLanguageKey: "en"},
		},
		sentMessages: []string{},
	}

	err := h.Handle(ctx)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	if len(ctx.sentMessages) == 0 {
		t.Fatal("Expected at least one message to be sent")
	}

	msg := ctx.sentMessages[0]
	t.Logf("Received message: %s", msg)

	// Expected order: Feed A (now), Feed C (yesterday), Feed B (twoDaysAgo), Feed D (nil)
	expectedOrder := []string{
		fmt.Sprintf("[[%d]] [%s](%s) - %s", 2, "Feed A", "http://example.com/a", now.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[[%d]] [%s](%s) - %s", 1, "Feed C", "http://example.com/c", yesterday.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[[%d]] [%s](%s) - %s", 3, "Feed B", "http://example.com/b", twoDaysAgo.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[[%d]] [%s](%s) - N/A", 4, "Feed D", "http://example.com/d"),
	}

	// Check if the message contains the sorted feeds
	for i, expectedLine := range expectedOrder {
		if !strings.Contains(msg, expectedLine) {
			t.Errorf("Message should contain line %d: %s", i, expectedLine)
		}
	}

	// Verify the order of lines in the message
	lines := strings.Split(msg, "\n")
	// Find where the actual feed list starts (after the header)
	startIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "[[") {
			startIdx = i
			break
		}
	}
	
	if startIdx >= 0 && len(lines) > startIdx {
		foundCount := 0
		for i := startIdx; i < len(lines) && foundCount < len(expectedOrder); i++ {
			line := strings.TrimSpace(lines[i])
			if line == "" {
				continue
			}
			// Check if this line matches the expected order
			if strings.Contains(line, fmt.Sprintf("[[%d]]", []int{2, 1, 3, 4}[foundCount])) {
				foundCount++
			}
		}
		if foundCount != len(expectedOrder) {
			t.Errorf("Expected to find all %d feeds in order, but found %d", len(expectedOrder), foundCount)
		}
	} else {
		t.Error("Could not find feed list in the message")
	}
}

func TestListSubscription_replaySubscribedSources_EmptyList(t *testing.T) {
	sourceStorage := &mockListSubSourceStorage{}
	subStorage := &mockListSubSubscriptionStorage{
		getUserSubsFunc: func(ctx context.Context, userID int64, opts *storage.GetSubscriptionsOptions) (*storage.GetSubscriptionsResult, error) {
			return &storage.GetSubscriptionsResult{
				Subscriptions: []*model.Subscribe{},
			}, nil
		},
	}

	coreInstance := core.NewCore(nil, nil, sourceStorage, subStorage, nil, nil)
	h := NewListSubscription(coreInstance)

	ctx := &mockListSubCtx{
		Context: &fakeContext{
			chatID: 123,
			data:   map[string]interface{}{util.UserLanguageKey: "en"},
		},
		sentMessages: []string{},
	}

	err := h.Handle(ctx)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	if len(ctx.sentMessages) == 0 {
		t.Fatal("Expected at least one message to be sent")
	}

	msg := ctx.sentMessages[0]
	// The actual translation key is "listsub_info_sub_list_empty" which translates to "Subscription list is empty."
	// But since we're not loading translations in the test, we get the translation missing message
	if !strings.Contains(msg, "listsub_info_sub_list_empty") {
		t.Errorf("Expected empty list message key, got: %s", msg)
	}
}

// Helper fake context for testing
type fakeContext struct {
	tb.Context
	chatID int64
	data   map[string]interface{}
}

func (f *fakeContext) Chat() *tb.Chat {
	return &tb.Chat{ID: f.chatID, Type: tb.ChatPrivate}
}

func (f *fakeContext) Get(key string) interface{} {
	return f.data[key]
}

func (f *fakeContext) Message() *tb.Message {
	return &tb.Message{Chat: &tb.Chat{ID: f.chatID}}
}

func (f *fakeContext) Bot() *tb.Bot {
	return &tb.Bot{}
}

func (f *fakeContext) Sender() *tb.User {
	return &tb.User{ID: 123}
}
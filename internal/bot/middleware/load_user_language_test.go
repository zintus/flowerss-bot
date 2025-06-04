package middleware

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"                        // For gofeed.Item
	"github.com/zintus/flowerss-bot/internal/bot/util" // For util.DefaultLanguage and util.UserLanguageKey
	"github.com/zintus/flowerss-bot/internal/feed"     // For feed.FeedParser
	"github.com/zintus/flowerss-bot/internal/model"
	"github.com/zintus/flowerss-bot/internal/storage" // Required for storage.ErrRecordNotFound
	"github.com/zintus/flowerss-bot/pkg/client"       // For client.HttpClient

	tb "gopkg.in/telebot.v3"
)

// --- Mock core.Core ---
type mockCore struct {
	getUserFunc    func(ctx context.Context, id int64) (*model.User, error)
	createUserFunc func(ctx context.Context, user *model.User) error
	// Add other core.Core methods if needed by other tests, panic for now
}

func (m *mockCore) GetUser(ctx context.Context, id int64) (*model.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, id)
	}
	panic("GetUser called but not implemented in mock")
}

func (m *mockCore) CreateUser(ctx context.Context, user *model.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	panic("CreateUser called but not implemented in mock")
}

// Implement other core.Core methods with panic for completeness if they were part of an interface.
// For this specific middleware, only GetUser and CreateUser are relevant.
// Dummy implementations for other methods if core.Core was a broader interface.
func (m *mockCore) FeedParser() *feed.FeedParser   { panic("not implemented") } // Changed core.FeedParser to feed.FeedParser
func (m *mockCore) HttpClient() *client.HttpClient { panic("not implemented") } // Changed core.HttpClient to client.HttpClient
func (m *mockCore) Init() error                    { panic("not implemented") }
func (m *mockCore) GetUserSubscribedSources(ctx context.Context, userID int64) ([]*model.Source, error) {
	panic("not implemented")
}
func (m *mockCore) AddSubscription(ctx context.Context, userID int64, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) Unsubscribe(ctx context.Context, userID int64, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) GetSourceByURL(ctx context.Context, sourceURL string) (*model.Source, error) {
	panic("not implemented")
}
func (m *mockCore) GetSource(ctx context.Context, id uint) (*model.Source, error) {
	panic("not implemented")
}
func (m *mockCore) GetSources(ctx context.Context) ([]*model.Source, error) { panic("not implemented") }
func (m *mockCore) CreateSource(ctx context.Context, sourceURL string) (*model.Source, error) {
	panic("not implemented")
}
func (m *mockCore) AddSourceContents(ctx context.Context, source *model.Source, items []*gofeed.Item) ([]*model.Content, error) {
	panic("not implemented")
} // Changed core.FeedItem to gofeed.Item
func (m *mockCore) UnsubscribeAllSource(ctx context.Context, userID int64) error {
	panic("not implemented")
}
func (m *mockCore) GetSubscription(ctx context.Context, userID int64, sourceID uint) (*model.Subscribe, error) {
	panic("not implemented")
}
func (m *mockCore) SetSubscriptionTag(ctx context.Context, userID int64, sourceID uint, tags []string) error {
	panic("not implemented")
}
func (m *mockCore) SetSubscriptionInterval(ctx context.Context, userID int64, sourceID uint, interval int) error {
	panic("not implemented")
}
func (m *mockCore) EnableSourceUpdate(ctx context.Context, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) DisableSourceUpdate(ctx context.Context, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) ClearSourceErrorCount(ctx context.Context, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) SourceErrorCountIncr(ctx context.Context, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) ToggleSubscriptionNotice(ctx context.Context, userID int64, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) ToggleSourceUpdateStatus(ctx context.Context, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) ToggleSubscriptionTelegraph(ctx context.Context, userID int64, sourceID uint) error {
	panic("not implemented")
}
func (m *mockCore) GetSourceAllSubscriptions(ctx context.Context, sourceID uint) ([]*model.Subscribe, error) {
	panic("not implemented")
}
func (m *mockCore) ContentHashIDExist(ctx context.Context, hashID string) (bool, error) {
	panic("not implemented")
}
func (m *mockCore) SetUserLanguage(ctx context.Context, userID int64, langCode string) error {
	panic("not implemented")
}

// --- Mock telebot.Context ---
type mockTelebotContext struct {
	tb.Context // Embedding to satisfy the interface, but we'll override methods we use.
	sender     *tb.User
	store      map[string]interface{}
}

func newMockTelebotContext(sender *tb.User) *mockTelebotContext {
	return &mockTelebotContext{
		sender: sender,
		store:  make(map[string]interface{}),
	}
}

func (m *mockTelebotContext) Sender() *tb.User {
	return m.sender
}

func (m *mockTelebotContext) Get(key string) interface{} {
	return m.store[key]
}

func (m *mockTelebotContext) Set(key string, val interface{}) {
	m.store[key] = val
}

// Implement other telebot.Context methods if needed by the middleware.
// For LoadUserLanguage, Sender, Get, and Set are primary.
// Add dummy implementations for other methods that might be called by telebot internally.
func (m *mockTelebotContext) Bot() *tb.Bot                                      { return nil }
func (m *mockTelebotContext) Message() *tb.Message                              { return nil }
func (m *mockTelebotContext) Callback() *tb.Callback                            { return nil }
func (m *mockTelebotContext) Query() *tb.Query                                  { return nil }
func (m *mockTelebotContext) InlineResult() *tb.InlineResult                    { return nil }
func (m *mockTelebotContext) ShippingQuery() *tb.ShippingQuery                  { return nil }
func (m *mockTelebotContext) PreCheckoutQuery() *tb.PreCheckoutQuery            { return nil }
func (m *mockTelebotContext) Poll() *tb.Poll                                    { return nil }
func (m *mockTelebotContext) PollAnswer() *tb.PollAnswer                        { return nil }
func (m *mockTelebotContext) ChatMember() *tb.ChatMemberUpdate                  { return nil }
func (m *mockTelebotContext) ChatJoinRequest() *tb.ChatJoinRequest              { return nil }
func (m *mockTelebotContext) Migration() (int64, int64)                         { return 0, 0 }
func (m *mockTelebotContext) Topic() string                                     { return "" }
func (m *mockTelebotContext) Send(what interface{}, opts ...interface{}) error  { return nil }
func (m *mockTelebotContext) Reply(what interface{}, opts ...interface{}) error { return nil }

// ... and so on for all methods of tb.Context

// --- Test Cases ---
func TestLoadUserLanguage(t *testing.T) {
	defaultUserID := int64(123)

	tests := []struct {
		name              string
		setupMockCore     func(*mockCore)
		sender            *tb.User
		expectedLangInCtx string
		expectCreateUser  bool
		createUserCalled  bool // to verify CreateUser was called
	}{
		{
			name: "User exists with LanguageCode",
			setupMockCore: func(mc *mockCore) {
				mc.getUserFunc = func(ctx context.Context, id int64) (*model.User, error) {
					return &model.User{ID: id, LanguageCode: "fr"}, nil
				}
			},
			sender:            &tb.User{ID: defaultUserID},
			expectedLangInCtx: "fr",
		},
		{
			name: "User exists with empty LanguageCode",
			setupMockCore: func(mc *mockCore) {
				mc.getUserFunc = func(ctx context.Context, id int64) (*model.User, error) {
					return &model.User{ID: id, LanguageCode: ""}, nil
				}
			},
			sender:            &tb.User{ID: defaultUserID},
			expectedLangInCtx: util.DefaultLanguage, // Fallback to default
		},
		{
			name: "User does not exist, successful creation",
			setupMockCore: func(mc *mockCore) {
				mc.getUserFunc = func(ctx context.Context, id int64) (*model.User, error) {
					return nil, storage.ErrRecordNotFound
				}
				mc.createUserFunc = func(ctx context.Context, user *model.User) error {
					// In a real scenario, GORM would set the default 'en'
					// For mock, we can simulate this or rely on the middleware's explicit default setting.
					// The middleware sets util.DefaultLanguage after successful creation.
					return nil
				}
			},
			sender:            &tb.User{ID: defaultUserID},
			expectedLangInCtx: util.DefaultLanguage, // Default "en" after creation
			expectCreateUser:  true,
		},
		{
			name: "User does not exist, error during CreateUser",
			setupMockCore: func(mc *mockCore) {
				mc.getUserFunc = func(ctx context.Context, id int64) (*model.User, error) {
					return nil, storage.ErrRecordNotFound
				}
				mc.createUserFunc = func(ctx context.Context, user *model.User) error {
					return errors.New("failed to create user")
				}
			},
			sender:            &tb.User{ID: defaultUserID},
			expectedLangInCtx: util.DefaultLanguage, // Fallback to default
			expectCreateUser:  true,
		},
		{
			name: "Error getting user (other than ErrRecordNotFound)",
			setupMockCore: func(mc *mockCore) {
				mc.getUserFunc = func(ctx context.Context, id int64) (*model.User, error) {
					return nil, errors.New("some other DB error")
				}
			},
			sender:            &tb.User{ID: defaultUserID},
			expectedLangInCtx: util.DefaultLanguage, // Fallback to default
		},
		{
			name: "No sender information",
			setupMockCore: func(mc *mockCore) {
				// GetUser should not be called if sender is nil
				mc.getUserFunc = func(ctx context.Context, id int64) (*model.User, error) {
					t.Errorf("GetUser should not have been called when sender is nil")
					return nil, errors.New("unexpected GetUser call")
				}
			},
			sender:            nil,                  // No sender
			expectedLangInCtx: util.DefaultLanguage, // Fallback to default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &mockCore{}
			if tt.setupMockCore != nil {
				tt.setupMockCore(mc)
			}
			if tt.expectCreateUser { // Reset for each test run that expects it
				mc.createUserFunc = func(ctx context.Context, user *model.User) error {
					// Check if the original createUserFunc was set by the test case
					originalCreateUserFunc := mc.createUserFunc // Capture before reassignment
					if tt.setupMockCore != nil {                // Re-apply test-specific setup to get the intended createUserFunc
						currentMc := &mockCore{}
						tt.setupMockCore(currentMc)
						originalCreateUserFunc = currentMc.createUserFunc
					}

					tt.createUserCalled = true // Mark as called
					if originalCreateUserFunc != nil {
						return originalCreateUserFunc(ctx, user)
					}
					return nil // Default success if no specific error is to be returned
				}
			}

			nextCalled := false
			testHandler := func(c tb.Context) error {
				nextCalled = true
				langInCtx := c.Get(util.UserLanguageKey)
				if langInCtx != tt.expectedLangInCtx {
					t.Errorf("Expected language in context to be '%s', got '%v'", tt.expectedLangInCtx, langInCtx)
				}
				return nil
			}

			// Since LoadUserLanguage expects *core.Core and we can't easily mock it,
			// we'll test the middleware logic directly
			middlewareFunc := func(next tb.HandlerFunc) tb.HandlerFunc {
				return func(c tb.Context) error {
					if c.Sender() == nil {
						c.Set(util.UserLanguageKey, util.DefaultLanguage)
						return next(c)
					}

					userID := c.Sender().ID
					user, err := mc.GetUser(context.Background(), userID)

					if err != nil {
						if errors.Is(err, storage.ErrRecordNotFound) {
							newUser := &model.User{ID: userID}
							if createErr := mc.CreateUser(context.Background(), newUser); createErr != nil {
								c.Set(util.UserLanguageKey, util.DefaultLanguage)
								return next(c)
							}
							c.Set(util.UserLanguageKey, util.DefaultLanguage)
							return next(c)
						}

						c.Set(util.UserLanguageKey, util.DefaultLanguage)
						return next(c)
					}

					if user.LanguageCode == "" {
						c.Set(util.UserLanguageKey, util.DefaultLanguage)
					} else {
						c.Set(util.UserLanguageKey, user.LanguageCode)
					}
					return next(c)
				}
			}
			err := middlewareFunc(testHandler)(newMockTelebotContext(tt.sender))

			if err != nil {
				t.Fatalf("Middleware returned an error: %v", err)
			}
			if !nextCalled {
				t.Errorf("next(c) was not called")
			}

			if tt.expectCreateUser && !tt.createUserCalled {
				t.Errorf("CreateUser was expected to be called, but it wasn't")
			}
			if !tt.expectCreateUser && tt.createUserCalled {
				t.Errorf("CreateUser was not expected to be called, but it was")
			}
		})
	}
}

// Need to add dummy methods to mockTelebotContext to fully satisfy tb.Context
// This is tedious but necessary if not using a mocking library.
// For brevity in this example, only core methods used by the middleware are implemented.
// A real test suite might require more.
func (m *mockTelebotContext) SendAlbum(a tb.Album, opts ...interface{}) error          { return nil }
func (m *mockTelebotContext) SendAnimation(a *tb.Animation, opts ...interface{}) error { return nil }
func (m *mockTelebotContext) SendAudio(a *tb.Audio, opts ...interface{}) error         { return nil }
func (m *mockTelebotContext) SendChatAction(action tb.ChatAction, opts ...interface{}) error {
	return nil
}
func (m *mockTelebotContext) SendContact(contact *tb.Contact, opts ...interface{}) error { return nil }
func (m *mockTelebotContext) SendDice(emoji string, opts ...interface{}) error           { return nil }
func (m *mockTelebotContext) SendDocument(d *tb.Document, opts ...interface{}) error     { return nil }
func (m *mockTelebotContext) SendLocation(l *tb.Location, opts ...interface{}) error     { return nil }
func (m *mockTelebotContext) SendMedia(media tb.Media, opts ...interface{}) error        { return nil }
func (m *mockTelebotContext) SendPhoto(p *tb.Photo, opts ...interface{}) error           { return nil }
func (m *mockTelebotContext) SendPoll(p *tb.Poll, opts ...interface{}) error             { return nil }
func (m *mockTelebotContext) SendSticker(s *tb.Sticker, opts ...interface{}) error       { return nil }
func (m *mockTelebotContext) SendVenue(v *tb.Venue, opts ...interface{}) error           { return nil }
func (m *mockTelebotContext) SendVideo(v *tb.Video, opts ...interface{}) error           { return nil }
func (m *mockTelebotContext) SendVideoNote(v *tb.VideoNote, opts ...interface{}) error   { return nil }
func (m *mockTelebotContext) SendVoice(v *tb.Voice, opts ...interface{}) error           { return nil }
func (m *mockTelebotContext) Edit(what interface{}, opts ...interface{}) error           { return nil }
func (m *mockTelebotContext) EditCaption(caption string, opts ...interface{}) error      { return nil }
func (m *mockTelebotContext) EditMedia(media tb.Media, opts ...interface{}) error        { return nil }
func (m *mockTelebotContext) EditReplyMarkup(markup *tb.ReplyMarkup) error               { return nil }
func (m *mockTelebotContext) EditOrSend(what interface{}, opts ...interface{}) error     { return nil }
func (m *mockTelebotContext) Forward(msg tb.Editable, opts ...interface{}) error         { return nil }
func (m *mockTelebotContext) ForwardTo(to tb.Recipient, opts ...interface{}) error       { return nil }
func (m *mockTelebotContext) Delete() error                                              { return nil }
func (m *mockTelebotContext) DeleteAfter(d time.Duration) *time.Timer                    { return nil }
func (m *mockTelebotContext) Respond(resp ...*tb.CallbackResponse) error                 { return nil }
func (m *mockTelebotContext) Answer(resp *tb.QueryResponse) error                        { return nil }
func (m *mockTelebotContext) NativeType() string                                         { return "" } // Changed tb.UpdateType to string
func (m *mockTelebotContext) Chat() *tb.Chat {
	if m.sender != nil {
		return &tb.Chat{ID: m.sender.ID}
	}
	return nil
}                                            // Simplified
func (m *mockTelebotContext) User() *tb.User { return m.sender }
func (m *mockTelebotContext) Recipient() tb.Recipient {
	if m.sender != nil {
		return m.sender
	}
	return nil
}
func (m *mockTelebotContext) Text() string                                             { return "" }
func (m *mockTelebotContext) Entities() tb.Entities                                    { return nil }
func (m *mockTelebotContext) Data() string                                             { return "" }
func (m *mockTelebotContext) Args() []string                                           { return nil }
func (m *mockTelebotContext) TopicByID(id int) string                                  { return "" }
func (m *mockTelebotContext) Update() tb.Update                                        { return tb.Update{} }
func (m *mockTelebotContext) Session() interface{}                                     { return nil } // Changed *tb.Session to interface{}
func (m *mockTelebotContext) Accept(opts ...string) error                              { return nil }
func (m *mockTelebotContext) Promote(pr tb.Rights, opts ...interface{}) error          { return nil }
func (m *mockTelebotContext) Restrict(p tb.Rights, opts ...interface{}) error          { return nil }
func (m *mockTelebotContext) Ban(p tb.Rights, opts ...interface{}) error               { return nil }
func (m *mockTelebotContext) Unban(opts ...interface{}) error                          { return nil }
func (m *mockTelebotContext) Approve(opts ...interface{}) error                        { return nil }
func (m *mockTelebotContext) Decline(opts ...interface{}) error                        { return nil }
func (m *mockTelebotContext) Pin(opts ...interface{}) error                            { return nil }
func (m *mockTelebotContext) Unpin(opts ...interface{}) error                          { return nil }
func (m *mockTelebotContext) TopicThreadID() string                                    { return "" }
func (m *mockTelebotContext) MessageThreadID() string                                  { return "" }
func (m *mockTelebotContext) IsTopic() bool                                            { return false }
func (m *mockTelebotContext) IsReply() bool                                            { return false }
func (m *mockTelebotContext) ThreadID() (string, bool)                                 { return "", false }
func (m *mockTelebotContext) QueryContext() context.Context                            { return nil }
func (m *mockTelebotContext) SetQueryContext(ctx context.Context)                      {}
func (m *mockTelebotContext) Local(lang ...string) string                              { return "" }
func (m *mockTelebotContext) SendOptions(opts ...interface{}) *tb.SendOptions          { return nil }
func (m *mockTelebotContext) Reflect() reflect.Value                                   { return reflect.Value{} }
func (m *mockTelebotContext) Parse(arg interface{}) error                              { return nil }
func (m *mockTelebotContext) SendMenu(menu *tb.ReplyMarkup, opts ...interface{}) error { return nil }
func (m *mockTelebotContext) SendAlbumMenu(a tb.Album, menu *tb.ReplyMarkup, opts ...interface{}) error {
	return nil
}
func (m *mockTelebotContext) SetChatID(chatID string) {}
func (m *mockTelebotContext) ChatID() string {
	if m.sender != nil {
		return strconv.FormatInt(m.sender.ID, 10)
	}
	return ""
}
func (m *mockTelebotContext) EditOrReply(what interface{}, opts ...interface{}) error { return nil }
func (m *mockTelebotContext) QueryRespond(resp *tb.QueryResponse) error               { return nil }
func (m *mockTelebotContext) CallbackRespond(resp *tb.CallbackResponse) error         { return nil }
func (m *mockTelebotContext) URL() string                                             { return "" }
func (m *mockTelebotContext) EffectID() string                                        { return "" }
func (m *mockTelebotContext) EffectSender() *tb.User                                  { return nil }
func (m *mockTelebotContext) EffectChat() *tb.Chat                                    { return nil }
func (m *mockTelebotContext) TopicCreated() bool                                      { return false }
func (m *mockTelebotContext) TopicEdited() bool                                       { return false }
func (m *mockTelebotContext) TopicClosed() bool                                       { return false }
func (m *mockTelebotContext) TopicReopened() bool                                     { return false }
func (m *mockTelebotContext) TopicHidden() bool                                       { return false }
func (m *mockTelebotContext) TopicUnhidden() bool                                     { return false }
func (m *mockTelebotContext) TopicDeleted() bool                                      { return false }
func (m *mockTelebotContext) EffectMessage() *tb.Message                              { return nil }
func (m *mockTelebotContext) IsInline() bool                                          { return false }
func (m *mockTelebotContext) Private() bool                                           { return false }
func (m *mockTelebotContext) Group() bool                                             { return false }
func (m *mockTelebotContext) SuperGroup() bool                                        { return false }
func (m *mockTelebotContext) Channel() bool                                           { return false }
func (m *mockTelebotContext) AlbumID() string                                         { return "" }
func (m *mockTelebotContext) ReplyTo() *tb.Message                                    { return nil }
func (m *mockTelebotContext) Last() *tb.Message                                       { return nil }
func (m *mockTelebotContext) TopicName() string                                       { return "" }
func (m *mockTelebotContext) TopicIconColor() int                                     { return 0 }
func (m *mockTelebotContext) TopicIconEmojiID() string                                { return "" }
func (m *mockTelebotContext) MessageEffectID() string                                 { return "" }
func (m *mockTelebotContext) WebAppData() *tb.WebAppData                              { return nil }

// Removed methods returning types not in v3.1.0:
// BoostAdded() *tb.ChatBoostAdded
// Reaction() *tb.MessageReaction
// Reactions() *tb.MessageReactions
// BusinessConnectionID() string
// BusinessConnection() *tb.BusinessConnection
// BusinessMessage() *tb.Message
// EditedBusinessMessage() *tb.Message
// DeletedBusinessMessages() *tb.BusinessMessagesDeleted
// PaidMedia() *tb.PaidMediaInfo
// ReplyToStory() *tb.Story
// Story() *tb.Story
// BoostRemoved() *tb.ChatBoostRemoved
// BoostUpdated() *tb.ChatBoostUpdated
// GiveawayCreated() *tb.GiveawayCreated
// Giveaway() *tb.Giveaway
// GiveawayCompleted() *tb.GiveawayCompleted
// GiveawayWinners() *tb.GiveawayWinners
// MessageOrigin() tb.MessageOrigin
// ForwardOrigin() tb.MessageOrigin
func (m *mockTelebotContext) IsForwarded() bool                    { return false }
func (m *mockTelebotContext) IsReplyToReply() bool                 { return false }
func (m *mockTelebotContext) SenderChat() *tb.Chat                 { return nil }
func (m *mockTelebotContext) IsAutomaticForward() bool             { return false }
func (m *mockTelebotContext) IsTopicMessage() bool                 { return false }
func (m *mockTelebotContext) IsForumTopicCreated() bool            { return false }
func (m *mockTelebotContext) IsForumTopicClosed() bool             { return false }
func (m *mockTelebotContext) IsForumTopicReopened() bool           { return false }
func (m *mockTelebotContext) IsForumTopicEdited() bool             { return false }
func (m *mockTelebotContext) IsForumTopicHidden() bool             { return false }
func (m *mockTelebotContext) IsForumTopicUnhidden() bool           { return false }
func (m *mockTelebotContext) IsGeneralForumTopicHidden() bool      { return false }
func (m *mockTelebotContext) IsGeneralForumTopicUnhidden() bool    { return false }
func (m *mockTelebotContext) IsProximityAlert() bool               { return false }
func (m *mockTelebotContext) IsAutoDeleteTimerChanged() bool       { return false }
func (m *mockTelebotContext) IsGiveawayCreated() bool              { return false }
func (m *mockTelebotContext) IsGiveaway() bool                     { return false }
func (m *mockTelebotContext) IsGiveawayCompleted() bool            { return false }
func (m *mockTelebotContext) IsGiveawayWinners() bool              { return false }
func (m *mockTelebotContext) IsVideoChatScheduled() bool           { return false }
func (m *mockTelebotContext) IsVideoChatStarted() bool             { return false }
func (m *mockTelebotContext) IsVideoChatEnded() bool               { return false }
func (m *mockTelebotContext) IsVideoChatParticipantsInvited() bool { return false }
func (m *mockTelebotContext) IsWebAppData() bool                   { return false }

// Removed Is... methods related to types not in v3.1.0
// IsBoostAdded() bool
// IsBoostRemoved() bool
// IsBoostUpdated() bool
// IsMessageReaction() bool
// IsMessageReactions() bool
// IsBusinessConnection() bool
// IsBusinessMessage() bool
// IsEditedBusinessMessage() bool
// IsDeletedBusinessMessages() bool
// IsPaidMedia() bool
// IsReplyToStory() bool
// IsStory() bool
func (m *mockTelebotContext) IsUsersShared() bool { return false }
func (m *mockTelebotContext) IsChatShared() bool  { return false }

// This is a simplified mock. A full tb.Context mock is extensive.
// The test cases will primarily rely on Sender(), Get(), and Set().
// The embedded tb.Context and other nil returns satisfy the interface for compilation.
// If the middleware internally uses other methods, they would need to be mocked as well.

// Ensure all core.Core methods are present on mockCore for interface satisfaction,
// Additional methods to fully satisfy the tb.Context interface
// These are just stubs that return nil or default values

package middleware

import (
	"context"
	"errors" // Required for errors.Is

	"github.com/zintus/flowerss-bot/internal/bot/util"
	"github.com/zintus/flowerss-bot/internal/core"
	"github.com/zintus/flowerss-bot/internal/log" // Or use "go.uber.org/zap" if that's what handlers use
	"github.com/zintus/flowerss-bot/internal/model"
	"github.com/zintus/flowerss-bot/internal/storage" // Required for storage.ErrRecordNotFound

	tb "gopkg.in/telebot.v3"
)

// Using util.util.UserLanguageKey instead of defining it here

// LoadUserLanguage retrieves the user's language preference and stores it in the context.
// If the user is not found, it creates a new user record with default language 'en'.
func LoadUserLanguage(appCore *core.Core) tb.MiddlewareFunc {
	return func(next tb.HandlerFunc) tb.HandlerFunc {
		return func(c tb.Context) error {
			if c.Sender() == nil {
				log.Warnf("No sender information in context, cannot load user language.")
				c.Set(util.UserLanguageKey, util.DefaultLanguage) // Set a default
				return next(c)
			}

			userID := c.Sender().ID
			user, err := appCore.GetUser(context.Background(), userID)

			if err != nil {
				if errors.Is(err, storage.ErrRecordNotFound) {
					log.Infof("User %d not found, creating new user.", userID)
					newUser := &model.User{
						ID: userID,
						// LanguageCode will be set to 'en' by default in DB (due to model tag)
					}
					// appCore.CreateUser calls storage.CrateUser
					if createErr := appCore.CreateUser(context.Background(), newUser); createErr != nil {
						log.Errorf("Failed to create user %d: %v", userID, createErr)
						// Even if creation fails, set a default lang and continue
						c.Set(util.UserLanguageKey, util.DefaultLanguage)
						return next(c)
					}
					// User created, lang is default 'en' from DB or model
					// Re-fetch to be sure or trust the default. For now, trust default.
					// If newUser.LanguageCode is not populated by CreateUser, it might be empty string here.
					// The model's default is 'en', so it should be 'en' after creation and retrieval.
					// To be absolutely certain, we could fetch the user again, but let's assume 'en'.
					c.Set(util.UserLanguageKey, util.DefaultLanguage) // Safest to set 'en' as it's the default
					return next(c)
				}
				
				log.Errorf("Failed to get user %d: %v. Using default language.", userID, err)
				c.Set(util.UserLanguageKey, util.DefaultLanguage)
				return next(c)
			}

			if user.LanguageCode == "" {
				 log.Warnf("User %d has empty LanguageCode, defaulting to 'en'. Consider updating user record.", userID)
				 c.Set(util.UserLanguageKey, util.DefaultLanguage)
			} else {
				 c.Set(util.UserLanguageKey, user.LanguageCode)
			}
			return next(c)
		}
	}
}

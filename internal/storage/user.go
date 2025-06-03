package storage

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/zintus/flowerss-bot/internal/model"
)

type UserStorageImpl struct {
	db *gorm.DB
}

func NewUserStorageImpl(db *gorm.DB) *UserStorageImpl {
	return &UserStorageImpl{db: db}
}

func (s *UserStorageImpl) Init(ctx context.Context) error {
	return s.db.Migrator().AutoMigrate(&model.User{})
}

func (s *UserStorageImpl) CreateUser(ctx context.Context, user *model.User) error {
	result := s.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *UserStorageImpl) GetUser(ctx context.Context, id int64) (*model.User, error) {
	var user = &model.User{}
	result := s.db.WithContext(ctx).Where(&model.User{ID: id}).First(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}
	return user, nil
}

func (s *UserStorageImpl) SetUserLanguage(ctx context.Context, userID int64, langCode string) error {
	result := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("language_code", langCode)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Assuming ErrRecordNotFound is already defined in this package or a common errors package.
		// If not, this might need adjustment or a generic error.
		return ErrRecordNotFound
	}
	return nil
}

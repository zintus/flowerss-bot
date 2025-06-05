package storage

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/zintus/flowerss-bot/internal/log"
	"github.com/zintus/flowerss-bot/internal/model"
)

type SourceStorageImpl struct {
	db *gorm.DB
}

func NewSourceStorageImpl(db *gorm.DB) *SourceStorageImpl {
	return &SourceStorageImpl{db: db}
}

func (s *SourceStorageImpl) Init(ctx context.Context) error {
	return s.db.Migrator().AutoMigrate(&model.Source{})
}

func (s *SourceStorageImpl) AddSource(ctx context.Context, source *model.Source) error {
	result := s.db.WithContext(ctx).Create(source)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SourceStorageImpl) GetSource(ctx context.Context, id uint) (*model.Source, error) {
	var source = &model.Source{}
	result := s.db.WithContext(ctx).Where(&model.Source{ID: id}).First(source)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}

	return source, nil
}

func (s *SourceStorageImpl) GetSources(ctx context.Context) ([]*model.Source, error) {
	var sources []*model.Source
	result := s.db.WithContext(ctx).Find(&sources)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}
	return sources, nil
}

func (s *SourceStorageImpl) GetSourceByURL(ctx context.Context, url string) (*model.Source, error) {
	var source = &model.Source{}
	result := s.db.WithContext(ctx).Where(&model.Source{Link: url}).First(source)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}
	return source, nil
}

func (s *SourceStorageImpl) Delete(ctx context.Context, id uint) error {
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Source{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SourceStorageImpl) UpsertSource(ctx context.Context, sourceID uint, newSource *model.Source) error {
	var oldSource model.Source
	result := s.db.WithContext(ctx).Where("id = ?", sourceID).First(&oldSource)

	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Source not found, create a new one
		newSource.ID = sourceID
		return s.db.WithContext(ctx).Create(newSource).Error
	} else {
		// Source found, update fields
		oldSource.Link = newSource.Link
		oldSource.Title = newSource.Title
		oldSource.ErrorCount = newSource.ErrorCount
		oldSource.LastPublishedAt = newSource.LastPublishedAt
		result = s.db.WithContext(ctx).Save(&oldSource)
		if result.Error != nil {
			return result.Error
		}
		log.Debugf("update %d row, sourceID %d new %#v", result.RowsAffected, sourceID, oldSource)
		return nil
	}
}

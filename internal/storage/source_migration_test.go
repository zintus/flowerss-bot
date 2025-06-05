package storage

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/zintus/flowerss-bot/internal/model"
)

func TestSourceStorage_Migration_AddsLastPublishedAt(t *testing.T) {
	// Create an in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create the sources table without LastPublishedAt (simulating old schema)
	// We'll manually create the table via SQL
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	// Create old schema without LastPublishedAt
	_, err = sqlDB.Exec(`
		CREATE TABLE sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			link TEXT,
			title TEXT,
			error_count INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create old schema: %v", err)
	}

	// Insert test data
	now := time.Now()
	_, err = sqlDB.Exec(`
		INSERT INTO sources (link, title, error_count, created_at, updated_at)
		VALUES 
			('http://example1.com/feed', 'Feed 1', 0, ?, ?),
			('http://example2.com/feed', 'Feed 2', 1, ?, ?),
			('http://example3.com/feed', 'Feed 3', 0, ?, ?)
	`, now, now, now, now, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Verify LastPublishedAt column doesn't exist
	var columnCount int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('sources')
		WHERE name = 'last_published_at'
	`).Scan(&columnCount)
	if err != nil {
		t.Fatalf("Failed to check column existence: %v", err)
	}
	if columnCount != 0 {
		t.Fatal("last_published_at column should not exist in old schema")
	}

	// Run the migration
	storage := NewSourceStorageImpl(db)
	if err := storage.Init(context.Background()); err != nil {
		t.Fatalf("Failed to run migration: %v", err)
	}

	// Verify LastPublishedAt column now exists
	err = sqlDB.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('sources')
		WHERE name = 'last_published_at'
	`).Scan(&columnCount)
	if err != nil {
		t.Fatalf("Failed to check column existence after migration: %v", err)
	}
	if columnCount != 1 {
		t.Fatal("last_published_at column should exist after migration")
	}

	// Verify we can query the data with the new model
	var sources []*model.Source
	if err := db.Find(&sources).Error; err != nil {
		t.Fatalf("Failed to query sources after migration: %v", err)
	}

	if len(sources) != 3 {
		t.Fatalf("Expected 3 sources, got %d", len(sources))
	}

	// Verify all existing sources have nil LastPublishedAt
	for _, src := range sources {
		if src.LastPublishedAt != nil {
			t.Errorf("Expected nil LastPublishedAt for migrated source %s, got %v", src.Title, src.LastPublishedAt)
		}
	}

	// Test updating LastPublishedAt
	updateTime := time.Now()
	sources[0].LastPublishedAt = &updateTime
	if err := db.Save(&sources[0]).Error; err != nil {
		t.Fatalf("Failed to update source with LastPublishedAt: %v", err)
	}

	// Verify the update
	var updatedSource model.Source
	if err := db.First(&updatedSource, sources[0].ID).Error; err != nil {
		t.Fatalf("Failed to query updated source: %v", err)
	}

	if updatedSource.LastPublishedAt == nil {
		t.Fatal("LastPublishedAt should not be nil after update")
	}

	if !updatedSource.LastPublishedAt.Equal(updateTime) {
		t.Errorf("Expected LastPublishedAt to be %v, got %v", updateTime, *updatedSource.LastPublishedAt)
	}
}

func TestSourceStorage_NewInstallation_HasLastPublishedAt(t *testing.T) {
	// Test that new installations create the table with LastPublishedAt field
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	storage := NewSourceStorageImpl(db)
	if err := storage.Init(context.Background()); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Verify the column exists
	sqlDB, _ := db.DB()
	var columnCount int
	err = sqlDB.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('sources')
		WHERE name = 'last_published_at'
	`).Scan(&columnCount)
	if err != nil {
		t.Fatalf("Failed to check column existence: %v", err)
	}
	if columnCount != 1 {
		t.Fatal("last_published_at column should exist in new installation")
	}

	// Insert a source with LastPublishedAt
	now := time.Now()
	source := &model.Source{
		Link:            "http://example.com/feed",
		Title:           "Test Feed",
		ErrorCount:      0,
		LastPublishedAt: &now,
	}

	if err := storage.AddSource(context.Background(), source); err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.GetSource(context.Background(), source.ID)
	if err != nil {
		t.Fatalf("Failed to get source: %v", err)
	}

	if retrieved.LastPublishedAt == nil {
		t.Fatal("LastPublishedAt should not be nil")
	}

	if !retrieved.LastPublishedAt.Equal(now) {
		t.Errorf("Expected LastPublishedAt to be %v, got %v", now, *retrieved.LastPublishedAt)
	}
}
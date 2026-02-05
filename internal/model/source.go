package model

import "time"

type Source struct {
	ID              uint `gorm:"primary_key;AUTO_INCREMENT"`
	Link            string
	Title           string
	ErrorCount      uint
	LastPublishedAt *time.Time
	LastContentAt   *time.Time // When we last received content locally (our timestamp, not from feed)
	Content         []Content
	EditTime
}

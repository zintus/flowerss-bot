package model

import "time"

type Source struct {
	ID              uint `gorm:"primary_key;AUTO_INCREMENT"`
	Link            string
	Title           string
	ErrorCount      uint
	LastPublishedAt *time.Time
	Content         []Content
	EditTime
}

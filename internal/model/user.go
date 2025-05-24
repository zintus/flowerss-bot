package model

// User subscriber
type User struct {
	ID           int64 `gorm:"primary_key"`
	LanguageCode string `gorm:"size:10;default:'en'"` // Added field
	EditTime
}

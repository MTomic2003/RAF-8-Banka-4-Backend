package model

import "gorm.io/gorm"

// WorkCode represents the industry classification (e.g. "10.1" = Food production)
type WorkCode struct {
	gorm.Model
	Code        string `gorm:"uniqueIndex;not null;size:10"` // "10.1"
	Description string `gorm:"not null"`
}

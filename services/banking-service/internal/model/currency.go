package model

import "gorm.io/gorm"

type Currency struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Code        string `gorm:"uniqueIndex;not null;size:4"` // EUR, RSD, USD...
	Symbol      string `gorm:"size:10"`
	Country     string
	Description string
	Status      string `gorm:"not null;default:'Active'"`
}

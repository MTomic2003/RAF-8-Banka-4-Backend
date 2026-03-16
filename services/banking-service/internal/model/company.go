package model

import "gorm.io/gorm"

type Company struct {
	gorm.Model
	Name               string `gorm:"not null"`
	RegistrationNumber string `gorm:"uniqueIndex;not null;size:8"`
	TaxNumber          string `gorm:"uniqueIndex;not null;size:9"`
	ActivityCodeID     uint
	WorkCode           *WorkCode `gorm:"foreignKey:ActivityCodeID"`
	Address            string
	OwnerID            uint `gorm:"not null"`
}

package model

type Company struct {
	CompanyID uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`

	RegistrationNumber string `gorm:"uniqueIndex;not null;size:8"`
	TaxNumber          string `gorm:"uniqueIndex;not null;size:9"`

	WorkCodeID uint `gorm:"index;not null;"`
	WorkCode   WorkCode

	Address string

	OwnerID uint `gorm:"not null"`
}

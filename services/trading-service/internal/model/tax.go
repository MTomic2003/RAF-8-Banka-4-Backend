package model

import (
	"time"
)

type TaxStatus string

const (
	TaxStatusCollected TaxStatus = "COLLECTED"
	TaxStatusFailed    TaxStatus = "FAILED"
)

type AccumulatedTax struct {
	AccumulatedTaxID uint    `gorm:"primaryKey"`
	AccountID        string  `gorm:"not null;uniqueIndex"`
	TaxOwedRSD       float64 `gorm:"not null;default:0"`
	LastUpdatedAt    time.Time
	LastClearedAt    *time.Time
}

type TaxCollection struct {
	TaxCollectionID   uint      `gorm:"primaryKey"`
	AccountID         string    `gorm:"not null"`
	TaxOwedRSD        float64   `gorm:"not null"`
	Status            TaxStatus `gorm:"type:varchar(20);not null"`
	FailureReason     string    `gorm:"type:text"`
	TaxingPeriodStart time.Time `gorm:"not null"`
	TaxingPeriodEnd   *time.Time
	TriggeredByID     *uint
}

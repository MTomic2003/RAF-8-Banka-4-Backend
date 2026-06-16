package model

import (
	"time"
)

type TaxStatus string

const (
	TaxStatusCollected TaxStatus = "COLLECTED"
	TaxStatusFailed    TaxStatus = "FAILED"
)

// AccumulatedTax holds the running capital-gains tax owed for an account. TaxOwed
// is always denominated in the account's own currency: RecordTax converts each
// taxable event into the account currency before accumulating, so no per-row
// currency needs to be stored.
type AccumulatedTax struct {
	AccumulatedTaxID uint    `gorm:"primaryKey;autoIncrement"`
	AccountNumber    string  `gorm:"not null;uniqueIndex:idx_acc_emp"`
	TaxOwed          float64 `gorm:"not null;default:0"`
	EmployeeID       *uint   `gorm:"uniqueIndex:idx_acc_emp"`
	LastUpdatedAt    time.Time
	LastClearedAt    *time.Time
}

// TaxCollection records the outcome of a single collection run. TaxOwed is in the
// account's currency (the same currency the collection payment is drawn in).
type TaxCollection struct {
	TaxCollectionID   uint    `gorm:"primaryKey;autoIncrement"`
	AccountNumber     string  `gorm:"not null"`
	TaxOwed           float64 `gorm:"not null"`
	EmployeeID        *uint
	Status            TaxStatus `gorm:"type:varchar(20);not null;check:status IN ('COLLECTED','FAILED')"`
	FailureReason     *string   `gorm:"type:text"`
	TaxingPeriodStart time.Time `gorm:"not null"`
	TaxingPeriodEnd   *time.Time
	TriggeredByID     *uint
}

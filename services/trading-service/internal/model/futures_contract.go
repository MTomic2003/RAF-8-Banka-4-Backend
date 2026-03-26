package model

import "time"

type FuturesContract struct {
	FuturesContractID uint      `gorm:"primaryKey;autoIncrement"`
	Ticker            string    `gorm:"not null;uniqueIndex;size:10"`
	Name              string    `gorm:"not null"`
	ContractSize      float64   `gorm:"not null"`
	ContractUnit      string    `gorm:"not null"`
	SettlementDate    time.Time `gorm:"not null"`
}

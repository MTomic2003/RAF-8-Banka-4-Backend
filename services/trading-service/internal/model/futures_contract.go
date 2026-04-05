package model

import "time"

type FuturesContract struct {
	FuturesContractID uint `gorm:"primaryKey;autoIncrement"`
	ListingID         uint `gorm:"not null;uniqueIndex"`
	Listing           Listing
	ContractSize      float64   `gorm:"not null"`
	ContractUnit      string    `gorm:"not null"`
	SettlementDate    time.Time `gorm:"not null"`
}

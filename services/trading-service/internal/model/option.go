package model

import "time"

type OptionType string

const (
	OptionTypeCall OptionType = "CALL"
	OptionTypePut  OptionType = "PUT"
)

type Option struct {
	OptionID          uint `gorm:"primaryKey;autoIncrement"`
	ListingID         uint `gorm:"not null;uniqueIndex"`
	Listing           Listing
	StockID           uint `gorm:"not null;"`
	Stock             Stock
	OptionType        OptionType `gorm:"not null;size:4"`
	StrikePrice       float64    `gorm:"not null;default:0"`
	ContractSize      int        `gorm:"not null;default:100"`
	SettlementDate    time.Time  `gorm:"not null"`
	ImpliedVolatility float64    `gorm:"not null;default:0"`
	OpenInterest      int        `gorm:"not null;default:0"`
}

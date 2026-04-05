package model

import (
	"time"
)

type ListingType string

const (
	ListingTypeStock     ListingType = "stock"
	ListingTypeOption    ListingType = "option"
	ListingTypeFuture    ListingType = "future"
	ListingTypeForexPair ListingType = "forexPair"
)

type Listing struct {
	ListingID         uint        `gorm:"primaryKey;autoIncrement"`
	Ticker            string      `gorm:"not null;uniqueIndex;size:20"`
	Name              string      `gorm:"not null"`
	ExchangeMIC       string      `gorm:"not null;size:10;index"`
	LastRefresh       time.Time   `gorm:"not null"`
	Price             float64     `gorm:"not null;default:0"`
	Ask               float64     `gorm:"not null;default:0"`
	MaintenanceMargin float64     `gorm:"not null;default:0"`
	ListingType       ListingType `gorm:"not null;size:10"`

	Exchange        *Exchange               `gorm:"foreignKey:ExchangeMIC;references:MicCode;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Stock           *Stock                  `gorm:"foreignKey:ListingID"`
	DailyPriceInfos []ListingDailyPriceInfo `gorm:"foreignKey:ListingID"`
}

type ListingDailyPriceInfo struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	ListingID uint `gorm:"not null;index"`
	Listing   Listing
	Date      time.Time `gorm:"not null;index"`
	Price     float64   `gorm:"not null;default:0"`
	Ask       float64   `gorm:"not null;default:0"`
	Bid       float64   `gorm:"not null;default:0"`
	Change    float64   `gorm:"not null;default:0"`
	Volume    uint      `gorm:"not null;default:0"`
}

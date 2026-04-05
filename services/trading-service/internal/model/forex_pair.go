package model

import "time"

type ForexPair struct {
	ForexPairID uint `gorm:"primaryKey"`

	ListingID uint `gorm:"not null;uniqueIndex"`
	Listing   Listing

	Base  string `gorm:"size:3;not null;uniqueIndex:idx_pair"`
	Quote string `gorm:"size:3;not null;uniqueIndex:idx_pair"`

	Rate float64 `gorm:"not null"`

	ProviderUpdatedAt    time.Time
	ProviderNextUpdateAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

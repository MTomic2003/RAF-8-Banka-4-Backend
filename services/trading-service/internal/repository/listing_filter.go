package repository

import "time"

type ListingFilter struct {
	Search         string
	Exchange       string
	PriceMin       float64
	PriceMax       float64
	AskMin         float64
	AskMax         float64
	BidMin         float64
	BidMax         float64
	VolumeMin      uint
	VolumeMax      uint
	SettlementDate *time.Time
	SortBy         string
	SortDir        string
	Page           int
	PageSize       int
}

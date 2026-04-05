package dto

import "time"

type ListingQuery struct {
	Search         string  `form:"search"`
	Exchange       string  `form:"exchange"`
	PriceMin       float64 `form:"price_min"`
	PriceMax       float64 `form:"price_max"`
	AskMin         float64 `form:"ask_min"`
	AskMax         float64 `form:"ask_max"`
	BidMin         float64 `form:"bid_min"`
	BidMax         float64 `form:"bid_max"`
	VolumeMin      uint    `form:"volume_min"`
	VolumeMax      uint    `form:"volume_max"`
	SettlementDate string  `form:"settlement_date"`
	SortBy         string  `form:"sort_by"`
	SortDir        string  `form:"sort_dir"`
	Page           int     `form:"page"`
	PageSize       int     `form:"page_size"`
}

func (q *ListingQuery) Normalize() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 10
	}
	if q.SortBy == "" {
		q.SortBy = "price"
	}
	if q.SortDir != "desc" {
		q.SortDir = "asc"
	}
}

func (q *ListingQuery) ParseSettlementDate() (*time.Time, error) {
	if q.SettlementDate == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", q.SettlementDate)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

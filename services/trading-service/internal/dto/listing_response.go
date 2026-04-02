package dto

import "time"

type BaseListingResponse struct {
	ListingID         uint    `json:"listing_id"`
	Ticker            string  `json:"ticker"`
	Name              string  `json:"name"`
	Exchange          string  `json:"exchange"`
	Price             float64 `json:"price"`
	Ask               float64 `json:"ask"`
	Bid               float64 `json:"bid"`
	Change            float64 `json:"change"`
	Volume            uint    `json:"volume"`
	MaintenanceMargin float64 `json:"maintenance_margin"`
	InitialMarginCost float64 `json:"initial_margin_cost"`
}

type StockResponse struct {
	BaseListingResponse
	OutstandingShares float64 `json:"outstanding_shares"`
	DividendYield     float64 `json:"dividend_yield"`
}

type FuturesResponse struct {
	BaseListingResponse
	SettlementDate time.Time `json:"settlement_date"`
	ContractSize   float64   `json:"contract_size"`
	ContractUnit   string    `json:"contract_unit"`
}

type ForexResponse struct {
	ForexPairID       uint    `json:"forexPairId"`
	Ticker            string  `json:"ticker"`
	Base              string  `json:"base"`
	Quote             string  `json:"quote"`
	Price             float64 `json:"price"`
	Ask               float64 `json:"ask"`
	Bid               float64 `json:"bid"`
	Change            float64 `json:"change"`
	Volume            uint    `json:"volume"`
	MaintenanceMargin float64 `json:"maintenanceMargin"`
	InitialMarginCost float64 `json:"initialMarginCost"`
}

type PaginatedResponse[T any] struct {
	Data     []T   `json:"data"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type OptionResponse struct {
	BaseListingResponse
	Strike            float64   `json:"strike"`
	OptionType        string    `json:"option_type"`
	SettlementDate    time.Time `json:"settlement_date"`
	ImpliedVolatility float64   `json:"implied_volatility"`
	OpenInterest      int       `json:"open_interest"`
}
type DailyPriceResponse struct {
	Date   time.Time `json:"date"`
	Price  float64   `json:"price"`
	Ask    float64   `json:"ask"`
	Bid    float64   `json:"bid"`
	Change float64   `json:"change"`
	Volume uint      `json:"volume"`
}

type StockDetailedResponse struct {
	StockResponse
	History []DailyPriceResponse `json:"history"`
	Options []OptionResponse     `json:"options"`
}

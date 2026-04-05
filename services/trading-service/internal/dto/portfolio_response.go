package dto

import "time"

type AssetType string

const (
	AssetTypeStock   AssetType = "STOCK"
	AssetTypeFutures AssetType = "FUTURES"
	AssetTypeOption  AssetType = "OPTION"
	AssetTypeForex   AssetType = "FOREX"
)

type PortfolioAssetResponse struct {
	Type              AssetType `json:"type"`
	Ticker            string    `json:"ticker"`
	Amount            float64   `json:"amount"`
	PricePerUnit      float64   `json:"pricePerUnit"`
	LastModified      time.Time `json:"lastModified"`
	Profit            float64   `json:"profit"`
	TaxAmount         float64   `json:"taxAmount"`
	OutstandingShares *float64  `json:"outstandingShares,omitempty"`
}

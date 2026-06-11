package dto

// CreatePriceAlertRequest is the body of POST /api/price-alerts. The owner of
// the alert is taken from the authenticated identity, not the request — the
// payload only describes WHAT to watch and WHEN to fire.
type CreatePriceAlertRequest struct {
	ListingID uint    `json:"listing_id" binding:"required"`
	Condition string  `json:"condition" binding:"required,oneof=ABOVE BELOW"`
	Threshold float64 `json:"threshold" binding:"required,gt=0"`
}

package dto

import "time"

type PriceAlertResponse struct {
	PriceAlertID     uint       `json:"price_alert_id"`
	ListingID        uint       `json:"listing_id"`
	Ticker           string     `json:"ticker,omitempty"`
	Condition        string     `json:"condition"`
	Threshold        float64    `json:"threshold"`
	NotificationType string     `json:"notification_type"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	TriggeredAt      *time.Time `json:"triggered_at,omitempty"`
}

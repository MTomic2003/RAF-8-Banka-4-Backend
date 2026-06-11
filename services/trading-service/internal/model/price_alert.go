package model

import "time"

// PriceAlertCondition is how the threshold is compared against the listing's
// current price when the scheduler ticks.
type PriceAlertCondition string

const (
	PriceAlertConditionAbove PriceAlertCondition = "ABOVE"
	PriceAlertConditionBelow PriceAlertCondition = "BELOW"
)

// PriceAlertNotificationType is the channel used to deliver the alert. The
// field exists to honour the specification, but the only channel implemented
// today is EMAIL — the value is set on every record and never branched on by
// the scheduler.
type PriceAlertNotificationType string

const PriceAlertNotificationEmail PriceAlertNotificationType = "EMAIL"

// PriceAlert is a personal threshold alarm: when the current price of a listing
// crosses the configured threshold in the chosen direction, the owner gets
// emailed once and the alert auto-deactivates (IsActive=false, TriggeredAt
// stamped). The user can create a new alert to be notified again.
type PriceAlert struct {
	PriceAlertID     uint                       `gorm:"primaryKey;autoIncrement"`
	UserID           uint                       `gorm:"not null;index"`
	OwnerType        OwnerType                  `gorm:"not null;size:10"`
	ListingID        uint                       `gorm:"not null;index"`
	Listing          *Listing                   `gorm:"foreignKey:ListingID;references:ListingID;constraint:-"`
	Condition        PriceAlertCondition        `gorm:"not null;size:10"`
	Threshold        float64                    `gorm:"not null"`
	NotificationType PriceAlertNotificationType `gorm:"not null;size:10;default:'EMAIL'"`
	IsActive         bool                       `gorm:"not null;default:true;index"`
	CreatedAt        time.Time                  `gorm:"not null"`
	TriggeredAt      *time.Time
}

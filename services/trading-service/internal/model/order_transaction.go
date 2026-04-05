package model

import "time"

type OrderTransaction struct {
	OrderTransactionID uint `gorm:"primaryKey;autoIncrement"`
	OrderID            uint `gorm:"not null;index"`
	Order              Order
	Quantity           uint      `gorm:"not null"`
	PricePerUnit       float64   `gorm:"not null"`
	TotalPrice         float64   `gorm:"not null"`
	Commission         float64   `gorm:"not null;default:0"`
	ExecutedAt         time.Time `gorm:"not null"`
	CreatedAt          time.Time
}

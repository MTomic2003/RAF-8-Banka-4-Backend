package model

type OwnerType string

const (
	OwnerTypeClient  OwnerType = "CLIENT"
	OwnerTypeActuary OwnerType = "ACTUARY"
)

type OrderOwnership struct {
	OrderOwnershipID uint `gorm:"primaryKey;autoIncrement"`
	OrderID          uint `gorm:"not null;index"`
	Order            Order
	IdentityID       uint      `gorm:"not null;index"`
	OwnerType        OwnerType `gorm:"not null;size:10"`
	AccountNumber    string    `gorm:"not null;size:18"`
}

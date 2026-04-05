package model

type Payment struct {
	PaymentID       uint `gorm:"primaryKey"`
	TransactionID   uint `gorm:"not null"`
	RecipientName   string
	ReferenceNumber string
	PaymentCode     string
	Purpose         string
	FailedAttempts  int `gorm:"not null;default:0"`

	Transaction Transaction
}

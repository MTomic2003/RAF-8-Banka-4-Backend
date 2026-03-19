package model

type Transfer struct {
	TransferID    uint `gorm:"primaryKey"`
	TransactionID uint `gorm:"not null;uniqueIndex"`
	ExchangeRate  *float64
	Commission    float64 `gorm:"not null;default:0"`

	Transaction Transaction
}

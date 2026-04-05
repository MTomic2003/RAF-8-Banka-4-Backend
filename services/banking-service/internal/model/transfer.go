package model

type Transfer struct {
	TransferID    uint `gorm:"primaryKey"`
	TransactionID uint `gorm:"not null;uniqueIndex"`
	ExchangeRate  *float64

	Transaction Transaction
}

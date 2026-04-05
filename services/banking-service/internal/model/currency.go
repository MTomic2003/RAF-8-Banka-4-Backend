package model

type CurrencyCode string

const (
	EUR CurrencyCode = "EUR"
	USD CurrencyCode = "USD"
	CHF CurrencyCode = "CHF"
	GBP CurrencyCode = "GBP"
	JPY CurrencyCode = "JPY"
	CAD CurrencyCode = "CAD"
	AUD CurrencyCode = "AUD"
	RSD CurrencyCode = "RSD"
)

var AllowedCurrencies = map[CurrencyCode]bool{
	EUR: true,
	USD: true,
	CHF: true,
	GBP: true,
	JPY: true,
	CAD: true,
	AUD: true,
	RSD: true,
}

var AllowedForeignCurrencies = map[CurrencyCode]bool{
	EUR: true,
	USD: true,
	CHF: true,
	GBP: true,
	JPY: true,
	CAD: true,
	AUD: true,
}

type Currency struct {
	CurrencyID  uint         `gorm:"primaryKey"`
	Name        string       `gorm:"not null"`
	Code        CurrencyCode `gorm:"uniqueIndex;not null;size:4"` // EUR, RSD, USD...
	Symbol      string       `gorm:"size:10"`
	Country     string
	Description string
	Status      string `gorm:"not null;default:'Active'"`
}

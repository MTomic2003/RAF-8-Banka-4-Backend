package model

const SimulatedExchangeMIC = "XSIM"

type Exchange struct {
	ExchangeID     uint   `gorm:"primaryKey;autoIncrement"`
	Name           string `gorm:"not null;size:100"`
	Acronym        string `gorm:"not null;size:20"`
	MicCode        string `gorm:"not null;uniqueIndex;size:10"`
	Polity         string `gorm:"not null;size:100"`
	Currency       string `gorm:"not null;size:50"`
	TimeZone       int    `gorm:"not null;default:0"`
	OpenTime       string `gorm:"not null;size:5"`
	CloseTime      string `gorm:"not null;size:5"`
	TradingEnabled bool   `gorm:"not null;default:true"`
}

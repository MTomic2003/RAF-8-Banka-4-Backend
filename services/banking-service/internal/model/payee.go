package model

type Payee struct{
    PayeeID       uint         `gorm:"primaryKey"`
    ClientID      uint         `gorm:"not null;index"`
    Name          string       `gorm:"not null"`
    AccountNumber string       `gorm:"not null"`
}
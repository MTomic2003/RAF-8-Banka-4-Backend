package model

import "time"

type Client struct {
	ClientID                 uint   `gorm:"primaryKey"`
	IdentityID               uint   `gorm:"uniqueIndex;not null"`
	FirstName                string `gorm:"size:20;not null"`
	LastName                 string `gorm:"size:100;not null"`
	MobileVerificationSecret string `gorm:"size:64"`
	DateOfBirth              time.Time
	Gender                   string `gorm:"size:10"`
	PhoneNumber              string `gorm:"size:20"`
	Address                  string `gorm:"size:255"`

	Identity Identity
}

package model

import (
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"time"
)

type Identity struct {
	ID           uint              `gorm:"primaryKey"`
	Email        string            `gorm:"size:100;uniqueIndex;not null"`
	Username     string            `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string            `gorm:"size:255"`
	Type         auth.IdentityType `gorm:"size:20;not null"`
	Active       bool              `gorm:"default:false"`

	LastFailedLoginTime time.Time
	FailedLoginCount    uint `gorm:"default:0"`

	ActivationTokens []ActivationToken `gorm:"foreignKey:IdentityID"`
	ResetTokens      []ResetToken      `gorm:"foreignKey:IdentityID"`
	RefreshTokens    []RefreshToken    `gorm:"foreignKey:IdentityID"`
}

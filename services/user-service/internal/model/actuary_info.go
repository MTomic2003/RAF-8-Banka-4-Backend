package model

import "time"

type ActuaryInfo struct {
	EmployeeID   uint    `gorm:"primaryKey"`
	IsAgent      bool    `gorm:"not null;default:false"`
	IsSupervisor bool    `gorm:"not null;default:false"`
	Limit        float64 `gorm:"type:numeric(20,2);not null;default:0"`
	UsedLimit    float64 `gorm:"type:numeric(20,2);not null;default:0"`
	NeedApproval bool    `gorm:"not null;default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (a ActuaryInfo) HasRole() bool {
	return a.IsAgent || a.IsSupervisor
}

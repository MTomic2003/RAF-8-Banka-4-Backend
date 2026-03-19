package model

import (
	"common/pkg/permission"
	"time"
)

type Employee struct {
	EmployeeID  uint   `gorm:"primaryKey"`
	IdentityID  uint   `gorm:"uniqueIndex;not null"`
	FirstName   string `gorm:"size:20;not null"`
	LastName    string `gorm:"size:100;not null"`
	Gender      string `gorm:"size:10"`
	DateOfBirth time.Time
	PhoneNumber string `gorm:"size:20"`
	Address     string `gorm:"size:255"`
	Department  string `gorm:"size:100"`
	PositionID  uint

	Identity    Identity
	Position    Position
	Permissions []EmployeePermission `gorm:"foreignKey:EmployeeID"`
}

func (e *Employee) HasPermission(p permission.Permission) bool {
	for _, ep := range e.Permissions {
		if ep.Permission == p {
			return true
		}
	}
	return false
}

func (e *Employee) RawPermissions() []permission.Permission {
	if e == nil || len(e.Permissions) == 0 {
		return []permission.Permission{}
	}

	permissions := make([]permission.Permission, 0, len(e.Permissions))
	for _, ep := range e.Permissions {
		permissions = append(permissions, ep.Permission)
	}

	return permissions
}

func (e *Employee) IsAdmin() bool {
	isAdmin := true
	for _, p := range permission.All {
		if !e.HasPermission(p) {
			isAdmin = false
		}
	}
	return isAdmin
}

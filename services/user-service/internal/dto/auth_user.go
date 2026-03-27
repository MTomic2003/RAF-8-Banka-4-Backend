package dto

import (
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type AuthUser struct {
	ID           uint                    `json:"id"`
	IdentityType auth.IdentityType       `json:"identity_type"`
	FirstName    string                  `json:"first_name"`
	LastName     string                  `json:"last_name"`
	Email        string                  `json:"email"`
	Username     string                  `json:"username"`
	Permissions  []permission.Permission `json:"permissions"`
	IsAgent      bool                    `json:"is_agent"`
	IsSupervisor bool                    `json:"is_supervisor"`
	Limit        float64                 `json:"limit"`
	UsedLimit    float64                 `json:"used_limit"`
	NeedApproval bool                    `json:"need_approval"`
}

func NewAuthUserFromEmployee(identity *model.Identity, employee *model.Employee) *AuthUser {
	return &AuthUser{
		ID:           employee.EmployeeID,
		IdentityType: identity.Type,
		FirstName:    employee.FirstName,
		LastName:     employee.LastName,
		Email:        identity.Email,
		Username:     identity.Username,
		Permissions:  employee.RawPermissions(),
		IsAgent:      employee.IsAgent(),
		IsSupervisor: employee.IsSupervisor(),
		Limit:        employeeLimit(employee),
		UsedLimit:    employeeUsedLimit(employee),
		NeedApproval: employeeNeedApproval(employee),
	}
}

func NewAuthUserFromClient(identity *model.Identity, client *model.Client) *AuthUser {
	return &AuthUser{
		ID:           client.ClientID,
		IdentityType: identity.Type,
		FirstName:    client.FirstName,
		LastName:     client.LastName,
		Email:        identity.Email,
		Username:     identity.Username,
		Permissions:  []permission.Permission{},
	}
}

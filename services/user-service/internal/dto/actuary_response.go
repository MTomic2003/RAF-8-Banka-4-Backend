package dto

import "github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"

type ActuaryResponse struct {
	ID           uint    `json:"id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	Username     string  `json:"username"`
	Department   string  `json:"department"`
	PositionID   uint    `json:"position_id"`
	Active       bool    `json:"active"`
	IsAgent      bool    `json:"is_agent"`
	IsSupervisor bool    `json:"is_supervisor"`
	Limit        float64 `json:"limit"`
	UsedLimit    float64 `json:"used_limit"`
	NeedApproval bool    `json:"need_approval"`
}

type ListActuariesResponse struct {
	Data       []ActuaryResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

func ToActuaryResponse(employee *model.Employee) *ActuaryResponse {
	response := &ActuaryResponse{
		ID:           employee.EmployeeID,
		FirstName:    employee.FirstName,
		LastName:     employee.LastName,
		Email:        employee.Identity.Email,
		Username:     employee.Identity.Username,
		Department:   employee.Department,
		PositionID:   employee.PositionID,
		Active:       employee.Identity.Active,
		IsAgent:      employee.IsAgent(),
		IsSupervisor: employee.IsSupervisor(),
	}

	if employee.ActuaryInfo != nil {
		response.Limit = employee.ActuaryInfo.Limit
		response.UsedLimit = employee.ActuaryInfo.UsedLimit
		response.NeedApproval = employee.ActuaryInfo.NeedApproval
	}

	return response
}

func ToActuaryResponseList(employees []model.Employee, total int64, page, pageSize int) *ListActuariesResponse {
	responses := make([]ActuaryResponse, len(employees))
	for i, employee := range employees {
		responses[i] = *ToActuaryResponse(&employee)
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &ListActuariesResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

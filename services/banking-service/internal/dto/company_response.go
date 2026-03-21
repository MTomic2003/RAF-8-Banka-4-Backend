package dto

import "github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"

type CompanyResponse struct {
	Name               string `json:"name" binding:"required"`
	RegistrationNumber string `json:"registration_number" binding:"required,max=8"`
	TaxNumber          string `json:"tax_number" binding:"required,max=9"`
	WorkCodeID         uint   `json:"work_code_id" binding:"required"`
	Address            string `json:"address"`
	OwnerID            uint   `json:"owner_id" binding:"required"`
}

func ToCompanyResponse(c *model.Company) CompanyResponse {
	return CompanyResponse{
		Name:               c.Name,
		RegistrationNumber: c.RegistrationNumber,
		TaxNumber:          c.TaxNumber,
		WorkCodeID:         c.WorkCodeID,
		Address:            c.Address,
		OwnerID:            c.OwnerID,
	}
}

package dto

import "github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"

type WorkCodeResponse struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func ToWorkCodeResponses(workCodes []model.WorkCode) []WorkCodeResponse {
	response := make([]WorkCodeResponse, 0, len(workCodes))
	for _, workCode := range workCodes {
		response = append(response, WorkCodeResponse{
			ID:          workCode.WorkCodeID,
			Code:        workCode.Code,
			Description: workCode.Description,
		})
	}

	return response
}

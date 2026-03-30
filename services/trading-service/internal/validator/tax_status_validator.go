package validator

import (
	"github.com/go-playground/validator/v10"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

func validateTaxStatus(fl validator.FieldLevel) bool {
	val := model.TaxStatus(fl.Field().String())
	return val == model.TaxStatusCollected || val == model.TaxStatusFailed
}

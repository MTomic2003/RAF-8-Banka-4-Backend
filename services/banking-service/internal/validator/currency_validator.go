package validator

import (
	"github.com/go-playground/validator/v10"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

func validateForeignCurrency(fl validator.FieldLevel) bool {
	return model.AllowedForeignCurrencies[model.CurrencyCode(fl.Field().String())]
}

package validator

import (
	"github.com/go-playground/validator/v10"
)

var allowedForeignCurrencies = map[string]bool{
	"EUR": true, "CHF": true, "USD": true,
	"GBP": true, "JPY": true, "CAD": true, "AUD": true,
}

func validateForeignCurrency(fl validator.FieldLevel) bool {
	return allowedForeignCurrencies[fl.Field().String()]
}

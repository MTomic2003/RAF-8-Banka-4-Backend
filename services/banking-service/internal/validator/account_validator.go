package validator

import (
	"github.com/go-playground/validator/v10"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

func validateAccountType(fl validator.FieldLevel) bool {
	val := model.AccountType(fl.Field().String())
	return val == model.AccountTypePersonal || val == model.AccountTypeBusiness
}

func validateAccountKind(fl validator.FieldLevel) bool {
	val := model.AccountKind(fl.Field().String())
	return val == model.AccountKindCurrent || val == model.AccountKindForeign
}

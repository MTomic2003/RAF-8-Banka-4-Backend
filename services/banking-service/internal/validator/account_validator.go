package validator

import (
	"banking-service/internal/model"

	"github.com/go-playground/validator/v10"
)

func validateAccountType(fl validator.FieldLevel) bool {
	val := model.AccountType(fl.Field().String())
	return val == model.AccountTypePersonal || val == model.AccountTypeBusiness
}

func validateAccountKind(fl validator.FieldLevel) bool {
	val := model.AccountKind(fl.Field().String())
	return val == model.AccountKindCurrent || val == model.AccountKindForeign
}

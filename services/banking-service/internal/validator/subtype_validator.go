package validator

import (
	"github.com/go-playground/validator/v10"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

func validateCurrentAccountStruct(sl validator.StructLevel) {
	req := sl.Current().Interface().(dto.CreateAccountRequest)

	if req.AccountKind != model.AccountKindCurrent {
		return
	}

	switch req.AccountType {
	case model.AccountTypePersonal:
		if !model.ValidPersonalSubtypes[req.Subtype] {
			sl.ReportError(req.Subtype, "Subtype", "subtype", "subtype_personal", "")
		}
	case model.AccountTypeBusiness:
		if !model.ValidBusinessSubtypes[req.Subtype] {
			sl.ReportError(req.Subtype, "Subtype", "subtype", "subtype_business", "")
		}
	}
}

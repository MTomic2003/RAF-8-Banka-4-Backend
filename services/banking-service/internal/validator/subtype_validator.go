package validator

import (
	"banking-service/internal/dto"
	"banking-service/internal/model"

	"github.com/go-playground/validator/v10"
)

var validPersonalSubtypes = map[model.Subtype]bool{
	model.SubtypeStandard:   true,
	model.SubtypeSavings:    true,
	model.SubtypePension:    true,
	model.SubtypeYouth:      true,
	model.SubtypeStudent:    true,
	model.SubtypeUnemployed: true,
}

var validBusinessSubtypes = map[model.Subtype]bool{
	model.SubtypeLLC:        true,
	model.SubtypeJointStock: true,
	model.SubtypeFoundation: true,
}

func validateCurrentAccountStruct(sl validator.StructLevel) {
	req := sl.Current().Interface().(dto.CreateAccountRequest)

	if req.AccountKind != model.AccountKindCurrent {
		return
	}

	switch req.AccountType {
	case model.AccountTypePersonal:
		if !validPersonalSubtypes[req.Subtype] {
			sl.ReportError(req.Subtype, "Subtype", "subtype", "subtype_personal", "")
		}
	case model.AccountTypeBusiness:
		if !validBusinessSubtypes[req.Subtype] {
			sl.ReportError(req.Subtype, "Subtype", "subtype", "subtype_business", "")
		}
	}
}

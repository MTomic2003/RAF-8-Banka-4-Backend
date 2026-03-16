package validator

import (
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
	accountTypeField := sl.Current().FieldByName("AccountType")
	subtypeField := sl.Current().FieldByName("Subtype")

	if !accountTypeField.IsValid() || !subtypeField.IsValid() {
		return
	}

	accountType := model.AccountType(accountTypeField.String())
	subtype := model.Subtype(subtypeField.String())

	switch accountType {
	case model.AccountTypePersonal:
		if !validPersonalSubtypes[subtype] {
			sl.ReportError(subtype, "Subtype", "subtype", "subtype_personal", "")
		}
	case model.AccountTypeBusiness:
		if !validBusinessSubtypes[subtype] {
			sl.ReportError(subtype, "Subtype", "subtype", "subtype_business", "")
		}
	}
}

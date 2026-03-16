package dto

import "banking-service/internal/model"

func ToAccountResponse(a *model.Account) AccountResponse {
	return AccountResponse{
		AccountNumber:    a.AccountNumber,
		Name:             a.Name,
		ClientID:         a.ClientID,
		CompanyID:        a.CompanyID,
		EmployeeID:       a.EmployeeID,
		Balance:          a.Balance,
		AvailableBalance: a.AvailableBalance,
		CreatedAt:        a.CreatedAt,
		ExpiresAt:        a.ExpiresAt,
		Status:           a.Status,
		AccountType:      string(a.AccountType),
		AccountKind:      string(a.AccountKind),
		Subtype:          string(a.Subtype),
		MaintenanceFee:   a.MaintenanceFee,
		DailyLimit:       a.DailyLimit,
		MonthlyLimit:     a.MonthlyLimit,
		DailySpending:    a.DailySpending,
		MonthlySpending:  a.MonthlySpending,
	}
}

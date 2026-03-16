package dto

import "time"

type AccountResponse struct {
	AccountNumber    string    `json:"account_number"`
	Name             string    `json:"name"`
	ClientID         uint      `json:"client_id"`
	CompanyID        *uint     `json:"company_id,omitempty"`
	EmployeeID       uint      `json:"employee_id"`
	Balance          float64   `json:"balance"`
	AvailableBalance float64   `json:"available_balance"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	//Currency         CurrencyResponse `json:"currency"`
	Status          string  `json:"status"`
	AccountType     string  `json:"account_type"`
	AccountKind     string  `json:"account_kind"`
	Subtype         string  `json:"subtype,omitempty"`
	MaintenanceFee  float64 `json:"maintenance_fee,omitempty"`
	DailyLimit      float64 `json:"daily_limit"`
	MonthlyLimit    float64 `json:"monthly_limit"`
	DailySpending   float64 `json:"daily_spending"`
	MonthlySpending float64 `json:"monthly_spending"`
}

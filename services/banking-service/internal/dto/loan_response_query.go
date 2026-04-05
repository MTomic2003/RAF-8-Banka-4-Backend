package dto

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type LoanRequestResponse struct {
	ID                 uint                    `json:"id"`
	ClientID           uint                    `json:"client_id"`
	AccountNumber      string                  `json:"account_number"`
	LoanType           string                  `json:"loan_type"`
	Amount             float64                 `json:"amount"`
	RepaymentPeriod    int                     `json:"repayment_period"`
	MonthlyInstallment float64                 `json:"monthly_installment"`
	Status             model.LoanRequestStatus `json:"status"`
	InstallmentDueDate *time.Time              `json:"installment_due_date,omitempty"`
}

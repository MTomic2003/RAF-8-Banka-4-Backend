package service

import "context"

type TaxRecorder interface {
	RecordTax(ctx context.Context, accountNumber string, employeeID *uint, profit float64, currencyCode string) error
}

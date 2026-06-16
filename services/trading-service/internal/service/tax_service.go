package service

import (
	"context"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/audit"
	commonauth "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/config"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

const taxRate = 0.15

type TaxService struct {
	taxRepo          repository.TaxRepository
	bankingClient    client.BankingClient
	taxAccountNumber string
	auditSvc         *audit.Service
}

func NewTaxService(
	taxRepo repository.TaxRepository,
	bankingClient client.BankingClient,
	cfg *config.Configuration,
	auditSvc *audit.Service,
) *TaxService {
	return &TaxService{
		taxRepo:          taxRepo,
		bankingClient:    bankingClient,
		taxAccountNumber: cfg.TaxAccountNumber,
		auditSvc:         auditSvc,
	}
}

func (s *TaxService) RecordTax(ctx context.Context, accountNumber string, employeeID *uint, profit float64, currencyCode string) error {
	if profit <= 0 {
		return nil
	}

	// Tax is accumulated and later collected in the account's own currency, so the
	// computed amount must be converted from the profit's currency before storing.
	taxAmount, err := s.toAccountCurrency(ctx, accountNumber, profit*taxRate, currencyCode)
	if err != nil {
		return err
	}

	if err := s.taxRepo.AddTaxOwed(ctx, accountNumber, employeeID, taxAmount); err != nil {
		return errors.InternalErr(err)
	}

	return nil
}

// toAccountCurrency converts amount from currencyCode into the account's own
// currency. Accumulated tax is always stored in the account currency, so every
// amount added to or subtracted from a row must pass through here first. The
// returned error is already wrapped for the caller to return directly.
func (s *TaxService) toAccountCurrency(ctx context.Context, accountNumber string, amount float64, currencyCode string) (float64, error) {
	accountCurrency, err := s.bankingClient.GetAccountCurrency(ctx, accountNumber)
	if err != nil {
		return 0, errors.InternalErr(err)
	}
	if currencyCode == accountCurrency {
		return amount, nil
	}
	converted, err := s.bankingClient.ConvertCurrency(ctx, amount, currencyCode, accountCurrency)
	if err != nil {
		return 0, errors.InternalErr(err)
	}
	return converted, nil
}

// ReduceTax lowers an account's accumulated capital-gains tax by 15% of a realized
// loss base (clamped at zero). A loss with no prior gains in the period has nothing
// to offset, so no row is created. The loss base is given in currencyCode and is
// converted to the account's currency, since the accumulated tax it offsets is
// stored in the account's currency.
func (s *TaxService) ReduceTax(ctx context.Context, accountNumber string, employeeID *uint, lossBase float64, currencyCode string) error {
	if lossBase <= 0 {
		return nil
	}

	reduction, err := s.toAccountCurrency(ctx, accountNumber, lossBase*taxRate, currencyCode)
	if err != nil {
		return err
	}

	if err := s.taxRepo.ReduceTaxOwed(ctx, accountNumber, employeeID, reduction); err != nil {
		return errors.InternalErr(err)
	}

	return nil
}

func (s *TaxService) CollectTaxes(ctx context.Context) error {
	taxes, err := s.taxRepo.FindAllPositiveAccumulatedTax(ctx)
	if err != nil {
		return errors.InternalErr(err)
	}
	now := time.Now()

	for _, tax := range taxes {
		amountToCollect := tax.TaxOwed

		collectionErr := s.collectSingleTax(ctx, tax.AccountNumber, amountToCollect)

		var status model.TaxStatus
		var failureReason *string
		if collectionErr != nil {
			status = model.TaxStatusFailed
			reason := collectionErr.Error()
			failureReason = &reason
		} else {
			status = model.TaxStatusCollected
		}

		collection := &model.TaxCollection{
			AccountNumber:     tax.AccountNumber,
			EmployeeID:        tax.EmployeeID,
			TaxOwed:           amountToCollect,
			Status:            status,
			FailureReason:     failureReason,
			TaxingPeriodStart: tax.LastUpdatedAt,
			TaxingPeriodEnd:   &now,
		}

		err = s.taxRepo.RecordCollectionResult(ctx, collection, collectionErr == nil, amountToCollect, now)
		if err != nil {
			return errors.InternalErr(err)
		}
	}

	if authCtx := commonauth.GetAuthFromContext(ctx); authCtx != nil && authCtx.EmployeeID != nil {
		if err := s.auditSvc.Log(ctx, audit.ActionTaxCollectionTriggered, *authCtx.EmployeeID, ""); err != nil {
			return errors.InternalErr(err)
		}
	}

	return nil
}

func (s *TaxService) collectSingleTax(ctx context.Context, accountNumber string, amount float64) error {
	_, err := s.bankingClient.CreatePaymentWithoutVerification(ctx, &pb.CreatePaymentRequest{
		PayerAccountNumber:     accountNumber,
		RecipientAccountNumber: s.taxAccountNumber,
		RecipientName:          "Republika Srbija",
		Amount:                 amount,
		PaymentCode:            "253",
		Purpose:                "Porez na kapitalnu dobit",
	})
	return err
}

func (s *TaxService) GetAccumulatedTax(ctx context.Context, accountNumber string) (*model.AccumulatedTax, error) {
	tax, err := s.taxRepo.FindAccumulatedTaxByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	return tax, nil
}

func (s *TaxService) GetTaxCollections(ctx context.Context, accountNumber string) ([]model.TaxCollection, error) {
	collections, err := s.taxRepo.FindTaxCollectionsByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	return collections, nil
}

func (s *TaxService) GetEmployeeTotalTax(ctx context.Context, employeeID uint) (float64, error) {
	taxes, err := s.taxRepo.FindAccumulatedTaxByEmployeeID(ctx, employeeID)
	if err != nil {
		return 0, errors.InternalErr(err)
	}

	return s.sumToRSD(ctx, taxes)
}

func (s *TaxService) GetClientTotalTax(ctx context.Context, clientID uint64) (float64, error) {
	accountsResp, err := s.bankingClient.GetAccountsByClientID(ctx, clientID)
	if err != nil {
		return 0, errors.InternalErr(err)
	}

	accountNumbers := make([]string, 0, len(accountsResp.Accounts))
	for _, acc := range accountsResp.Accounts {
		accountNumbers = append(accountNumbers, acc.AccountNumber)
	}

	taxes, err := s.taxRepo.FindAccumulatedTaxByClientAccountNumbers(ctx, accountNumbers)
	if err != nil {
		return 0, errors.InternalErr(err)
	}

	return s.sumToRSD(ctx, taxes)
}

// sumToRSD totals the owed tax across accumulated-tax rows, converting each row
// from its account's currency to RSD. The per-row currency is resolved from the
// account (it is no longer stored on the row), since TaxOwed is always denominated
// in the account's own currency.
func (s *TaxService) sumToRSD(ctx context.Context, taxes []model.AccumulatedTax) (float64, error) {
	total := 0.0
	for _, t := range taxes {
		if t.TaxOwed <= 0 {
			continue
		}
		currency, err := s.bankingClient.GetAccountCurrency(ctx, t.AccountNumber)
		if err != nil {
			return 0, errors.InternalErr(err)
		}
		if currency == "RSD" {
			total += t.TaxOwed
			continue
		}
		converted, err := s.bankingClient.ConvertCurrency(ctx, t.TaxOwed, currency, "RSD")
		if err != nil {
			return 0, errors.InternalErr(err)
		}
		total += converted
	}
	return total, nil
}

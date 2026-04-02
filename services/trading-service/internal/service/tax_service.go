package service

import (
	"context"
	"time"

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
}

func NewTaxService(
	taxRepo repository.TaxRepository,
	bankingClient client.BankingClient,
	cfg *config.Configuration,
) *TaxService {
	return &TaxService{
		taxRepo:          taxRepo,
		bankingClient:    bankingClient,
		taxAccountNumber: cfg.TaxAccountNumber,
	}
}

// profit pozitivan, izračunava 15% poreza i upisuje u bazu.
func (s *TaxService) RecordTax(ctx context.Context, accountNumber string, profit float64) error {
	if profit <= 0 {
		return nil
	}

	taxAmount := profit * taxRate

	err := s.taxRepo.AddTaxOwed(ctx, accountNumber, taxAmount)
	if err != nil {
		return errors.InternalErr(err)
	}

	return nil
}

// Za svaki racun koji ima nakupljeni porez, pokušava da ga uplati na racun drzave, rezultat toga (uspeh ili neuspeh) u bazu.
func (s *TaxService) CollectTaxes(ctx context.Context) error {
	taxes, err := s.taxRepo.FindAllPositiveAccumulatedTax(ctx)
	if err != nil {
		return errors.InternalErr(err)
	}
	now := time.Now()

	for _, tax := range taxes {
		amountToCollect := tax.TaxOwedRSD

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
			TaxOwedRSD:        amountToCollect,
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

// trenutni nakupljeni porez za jedan racun iz baze
func (s *TaxService) GetAccumulatedTax(ctx context.Context, accountNumber string) (*model.AccumulatedTax, error) {
	tax, err := s.taxRepo.FindAccumulatedTaxByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	return tax, nil
}

// istoriju svih naplatnih pokusaja poreza racuna(uspeni ili ne)
func (s *TaxService) GetTaxCollections(ctx context.Context, accountNumber string) ([]model.TaxCollection, error) {
	collections, err := s.taxRepo.FindTaxCollectionsByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	return collections, nil
}

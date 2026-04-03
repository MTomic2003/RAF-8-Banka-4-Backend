package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/repository"
)

type LoanService struct {
	accountRepo     repository.AccountRepository
	loanTypeRepo    repository.LoanTypeRepository
	loanRepo        repository.LoanRepository
	loanRequestRepo repository.LoanRequestRepository
	txProcessor     *TransactionProcessor
	txManager       repository.TransactionManager
	userClient      client.UserClient
	mailer          Mailer
}

func NewLoanService(
	accountRepo repository.AccountRepository,
	loanTypeRepo repository.LoanTypeRepository,
	loanRequestRepo repository.LoanRequestRepository,
	loanRepo repository.LoanRepository,
	txProcessor *TransactionProcessor,
	txManager repository.TransactionManager,
	userClient client.UserClient,
	mailer Mailer,
) *LoanService {
	return &LoanService{
		accountRepo:     accountRepo,
		loanTypeRepo:    loanTypeRepo,
		loanRequestRepo: loanRequestRepo,
		loanRepo:        loanRepo,
		txProcessor:     txProcessor,
		txManager:       txManager,
		userClient:      userClient,
		mailer:          mailer,
	}
}

func (s *LoanService) CalculateMonthlyInstallment(amount float64, annualRatePercent float64, months int) float64 {
	if annualRatePercent <= 0 {
		if months == 0 {
			return 0
		}
		return amount / float64(months)
	}

	monthlyRate := (annualRatePercent / 100.0) / 12.0
	compoundFactor := math.Pow(1.0+monthlyRate, float64(months))
	installment := amount * (monthlyRate * compoundFactor) / (compoundFactor - 1.0)

	return math.Round(installment*100) / 100
}

func (s *LoanService) SubmitLoanRequest(ctx context.Context, req *dto.CreateLoanRequest, clientID uint) (*dto.CreateLoanResponse, error) {
	account, err := s.accountRepo.FindByAccountNumber(ctx, req.AccountNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if account == nil {
		return nil, errors.BadRequestErr("account not found")
	}

	loanType, err := s.loanTypeRepo.FindByID(ctx, req.LoanTypeID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if loanType == nil {
		return nil, errors.BadRequestErr("credit type not found")
	}

	if req.RepaymentPeriod < loanType.MinRepaymentPeriod || req.RepaymentPeriod > loanType.MaxRepaymentPeriod {
		return nil, errors.BadRequestErr("repayment period is not valid for loan type")
	}

	client, err := s.userClient.GetClientByID(ctx, clientID)
	if err != nil {
		return nil, errors.ServiceUnavailableErr(err)
	}
	if client == nil {
		return nil, errors.BadRequestErr("client not found in user service")
	}

	// RAČUNANJE KAMATE I RATE
	totalInterestRate := loanType.BaseInterestRate + loanType.BankMargin
	monthlyInstallment := s.CalculateMonthlyInstallment(req.Amount, totalInterestRate, req.RepaymentPeriod)

	newRequest := &model.LoanRequest{
		ClientID:           clientID,
		AccountNumber:      req.AccountNumber,
		LoanTypeID:         req.LoanTypeID,
		Amount:             req.Amount,
		RepaymentPeriod:    req.RepaymentPeriod,
		CalculatedRate:     totalInterestRate,
		MonthlyInstallment: monthlyInstallment,
		Status:             model.LoanRequestPending, // Kreira se sa statusom PENDING, kako piše u tasku
	}

	if err := s.loanRequestRepo.CreateRequest(ctx, newRequest); err != nil {
		return nil, errors.InternalErr(err)
	}

	if err := s.mailer.Send(client.Email, "Loan submitted", "Your loan request has been succesfully submitted."); err != nil {
		log.Printf("failed to send loan request confirmation email to client_id=%d: %v", client.Id, err)
	}

	return &dto.CreateLoanResponse{
		RequestID:          newRequest.ID,
		Status:             newRequest.Status,
		MonthlyInstallment: monthlyInstallment,
	}, nil
}

func (s *LoanService) GetClientLoans(ctx context.Context, clientID uint, sortByAmountDesc bool) ([]dto.LoanResponse, error) {
	loans, err := s.loanRepo.FindByClientID(ctx, clientID, sortByAmountDesc)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	var response []dto.LoanResponse
	for _, l := range loans {
		account, err := s.accountRepo.FindByAccountNumber(ctx, l.LoanRequest.AccountNumber)
		if err != nil {
			return nil, errors.InternalErr(err)
		}

		response = append(response, dto.LoanResponse{
			ID:                 l.ID,
			LoanType:           l.LoanRequest.LoanType.Name,
			Amount:             l.LoanRequest.Amount,
			Currency:           account.Currency.Code,
			MonthlyInstallment: l.MonthlyInstallment,
			Status:             l.LoanRequest.Status,
		})
	}
	return response, nil
}

func (s *LoanService) GetLoanDetails(ctx context.Context, clientID uint, loanID uint) (*dto.LoanDetailsResponse, error) {
	loan, err := s.loanRepo.FindByIDAndClientID(ctx, loanID, clientID)
	if err != nil {
		return nil, errors.NotFoundErr("loan not found")
	}

	var installments []dto.Installment
	for i := 1; i <= loan.RepaymentPeriod; i++ {
		installments = append(installments, dto.Installment{
			Number: i,
			Amount: loan.MonthlyInstallment,
			Status: "UPCOMING",
		})
	}

	account, err := s.accountRepo.FindByAccountNumber(ctx, loan.LoanRequest.AccountNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	return &dto.LoanDetailsResponse{
		LoanResponse: dto.LoanResponse{
			ID:                 loan.ID,
			LoanType:           loan.LoanRequest.LoanType.Name,
			Amount:             loan.LoanRequest.Amount,
			Currency:           account.Currency.Code,
			MonthlyInstallment: loan.MonthlyInstallment,
			Status:             loan.LoanRequest.Status,
		},
		RepaymentPeriod: loan.RepaymentPeriod,
		InterestRate:    loan.InterestRate,
		Installments:    installments,
	}, nil
}

func (s *LoanService) GetLoanRequests(ctx context.Context, query *dto.ListLoanRequestsQuery) ([]dto.LoanRequestResponse, int64, error) {
	requests, total, err := s.loanRequestRepo.FindAll(ctx, query)
	if err != nil {
		return nil, 0, errors.InternalErr(err)
	}

	var response []dto.LoanRequestResponse
	for _, r := range requests {
		var installmentDueDate *time.Time

		if r.Status == model.LoanRequestApproved {
			loan, err := s.loanRepo.FindLoanByRequestID(ctx, r.ID)
			if err == nil && loan != nil {
				installmentDueDate = &loan.NextInstallmentDate
			}
		}

		response = append(response, dto.LoanRequestResponse{
			ID:                 r.ID,
			ClientID:           r.ClientID,
			AccountNumber:      r.AccountNumber,
			LoanType:           r.LoanType.Name,
			Amount:             r.Amount,
			RepaymentPeriod:    r.RepaymentPeriod,
			MonthlyInstallment: r.MonthlyInstallment,
			Status:             r.Status,
			InstallmentDueDate: installmentDueDate,
		})
	}

	return response, total, nil
}

func (s *LoanService) ApproveLoanRequest(ctx context.Context, id uint) error {
	request, err := s.loanRequestRepo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalErr(err)
	}

	if request == nil {
		return errors.NotFoundErr("loan request not found")
	}

	if request.Status != model.LoanRequestPending {
		return errors.BadRequestErr("loan request is not pending")
	}

	clientAccount, err := s.accountRepo.FindByAccountNumber(ctx, request.AccountNumber)
	if err != nil {
		return errors.InternalErr(err)
	}
	if clientAccount == nil {
		return errors.BadRequestErr("client account not found")
	}

	bankAccountNumber, ok := BankAccounts[clientAccount.Currency.Code]
	if !ok {
		return errors.BadRequestErr("no bank account for this currency")
	}

	bankAccount, err := s.accountRepo.FindByAccountNumber(ctx, bankAccountNumber)
	if err != nil {
		return errors.InternalErr(err)
	}
	if bankAccount == nil {
		return errors.InternalErr(errors.BadRequestErr("bank account not found"))
	}

	if bankAccount.AvailableBalance < request.Amount {
		return errors.BadRequestErr("insufficient bank funds to approve loan")
	}

	now := time.Now()
	firstInstallmentDate := time.Date(now.Year(), now.Month()+1, now.Day(), 0, 0, 0, 0, time.UTC)

	installments := make([]model.LoanInstallment, request.RepaymentPeriod)
	for i := 0; i < request.RepaymentPeriod; i++ {
		dueDate := time.Date(now.Year(), now.Month()+time.Month(i+1), now.Day(), 0, 0, 0, 0, time.UTC)
		installments[i] = model.LoanInstallment{
			InstallmentNumber: i + 1,
			Amount:            request.MonthlyInstallment,
			InterestRate:      request.CalculatedRate,
			DueDate:           dueDate,
			Status:            model.InstallmentStatusPending,
		}
	}

	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		transaction := &model.Transaction{
			PayerAccountNumber:     bankAccountNumber,
			RecipientAccountNumber: request.AccountNumber,
			StartAmount:            request.Amount,
			StartCurrencyCode:      clientAccount.Currency.Code,
			EndAmount:              request.Amount,
			EndCurrencyCode:        clientAccount.Currency.Code,
			Status:                 model.TransactionProcessing,
		}
		if err := s.txProcessor.transactionRepo.Create(txCtx, transaction); err != nil {
			return errors.InternalErr(err)
		}

		if err := s.txProcessor.Process(txCtx, transaction.TransactionID); err != nil {
			return err
		}

		request.Status = model.LoanRequestApproved
		if err := s.loanRequestRepo.Update(txCtx, request); err != nil {
			return errors.InternalErr(err)
		}

		loan := &model.Loan{
			LoanRequestID:       request.ID,
			TransactionID:       &transaction.TransactionID,
			MonthlyInstallment:  request.MonthlyInstallment,
			InterestRate:        request.CalculatedRate,
			IsVariableRate:      false,
			RepaymentPeriod:     request.RepaymentPeriod,
			PaidInstallments:    0,
			StartDate:           now,
			NextInstallmentDate: firstInstallmentDate,
			Status:              model.LoanStatusActive,
		}
		if err := s.loanRepo.CreateLoan(txCtx, loan); err != nil {
			return errors.InternalErr(err)
		}

		for i := range installments {
			installments[i].LoanID = loan.ID
		}

		if err := s.loanRepo.CreateInstallments(txCtx, installments); err != nil {
			return errors.InternalErr(err)
		}

		return nil
	}); err != nil {
		return err
	}

	client, err := s.userClient.GetClientByID(ctx, request.ClientID)
	if err != nil {
		return errors.ServiceUnavailableErr(err)
	}
	if client == nil {
		return errors.InternalErr(fmt.Errorf("client not found in user service"))
	}

	if err := s.mailer.Send(client.Email, "Loan request approved", "Your loan request has been approved."); err != nil {
		log.Printf("failed to send loan approval notification email to client_id=%d: %v", client.Id, err)
	}

	return nil
}

func (s *LoanService) RejectLoanRequest(ctx context.Context, id uint) error {
	request, err := s.loanRequestRepo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalErr(err)
	}

	if request == nil {
		return errors.NotFoundErr("loan request not found")
	}

	if request.Status != model.LoanRequestPending {
		return errors.BadRequestErr("loan request is not pending")
	}

	client, err := s.userClient.GetClientByID(ctx, request.ClientID)
	if err != nil {
		return errors.ServiceUnavailableErr(err)
	}
	if client == nil {
		return errors.InternalErr(fmt.Errorf("client not found in user service"))
	}

	if err := s.mailer.Send(client.Email, "Loan request rejected", "Your loan request has been rejected."); err != nil {
		log.Printf("failed to send loan rejection notification email to client_id=%d: %v", client.Id, err)
	}

	request.Status = model.LoanRequestRejected
	return s.loanRequestRepo.Update(ctx, request)
}

// AdjustVariableRates mesecno azurira kamatnu stopu za varijabilne kredite.
// Poziva se iz LoanScheduler-a jednom mesecno.
func (s *LoanService) AdjustVariableRates(ctx context.Context) error {
	loans, err := s.loanRepo.FindActiveVariableRateLoans(ctx)
	if err != nil {
		return errors.InternalErr(err)
	}

	for i := range loans {
		loan := &loans[i]

		// slucajna promena u opsegu [-1.5%, +1.5%]
		adjustment := rand.Float64()*3.0 - 1.5
		newRate := math.Round((loan.InterestRate+adjustment)*100) / 100
		if newRate < 0 {
			newRate = 0
		}

		// preracunavamo ratu za preostali period otplate
		remaining := loan.RepaymentPeriod - loan.PaidInstallments
		newInstallment := s.CalculateMonthlyInstallment(loan.RemainingDebt, newRate, remaining)

		loan.InterestRate = newRate
		loan.MonthlyInstallment = newInstallment

		if err := s.loanRepo.UpdateLoan(ctx, loan); err != nil {
			continue
		}

		// azuriramo iznos i kamatu za sve buducе rate
		for j := range loan.Installments {
			inst := &loan.Installments[j]
			if inst.Status == model.InstallmentStatusPending || inst.Status == model.InstallmentStatusRetrying {
				inst.Amount = newInstallment
				inst.InterestRate = newRate
				_ = s.loanRepo.UpdateInstallment(ctx, inst)
			}
		}
	}

	return nil
}

package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/repository"
)

type fakeMailer struct{}

func (f *fakeMailer) Send(to, subject, body string) error {
	return nil
}

// ── Fake Loan Request Repository ─────────────────────────────────────────────

type fakeLoanRequestRepo struct {
	request    *model.LoanRequest
	requests   []model.LoanRequest
	total      int64
	createErr  error
	findErr    error
	findAllErr error
	updateErr  error
	updated    *model.LoanRequest
}

func (f *fakeLoanRequestRepo) CreateRequest(_ context.Context, r *model.LoanRequest) error {
	if f.createErr != nil {
		return f.createErr
	}
	r.ID = 1
	return nil
}

func (f *fakeLoanRequestRepo) FindAll(_ context.Context, _ *dto.ListLoanRequestsQuery) ([]model.LoanRequest, int64, error) {
	return f.requests, f.total, f.findAllErr
}

func (f *fakeLoanRequestRepo) FindByID(_ context.Context, _ uint) (*model.LoanRequest, error) {
	return f.request, f.findErr
}

func (f *fakeLoanRequestRepo) Update(_ context.Context, r *model.LoanRequest) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = r
	return nil
}

// ── Fake Loan Repository ─────────────────────────────────────────────────────

type fakeLoanRepo struct {
	loan      *model.Loan
	loans     []model.Loan
	loanErr   error
	instErr   error
	findErr   error
	updateErr error
}

func (f *fakeLoanRepo) FindByClientID(_ context.Context, _ uint, _ bool) ([]model.Loan, error) {
	return f.loans, f.findErr
}

func (f *fakeLoanRepo) FindByIDAndClientID(_ context.Context, _ uint, _ uint) (*model.Loan, error) {
	return f.loan, f.findErr
}

func (f *fakeLoanRepo) CreateLoan(_ context.Context, loan *model.Loan) error {
	if f.loanErr != nil {
		return f.loanErr
	}
	loan.ID = 1
	return nil
}

func (f *fakeLoanRepo) FindLoanByRequestID(_ context.Context, _ uint) (*model.Loan, error) {
	return nil, nil
}

func (f *fakeLoanRepo) UpdateLoan(_ context.Context, _ *model.Loan) error {
	return f.loanErr
}

func (f *fakeLoanRepo) CreateInstallments(_ context.Context, _ []model.LoanInstallment) error {
	return f.instErr
}

func (f *fakeLoanRepo) FindDueInstallments(_ context.Context, _ time.Time) ([]model.LoanInstallment, error) {
	return nil, nil
}

func (f *fakeLoanRepo) FindRetryInstallments(_ context.Context, _ time.Time) ([]model.LoanInstallment, error) {
	return nil, nil
}

func (f *fakeLoanRepo) UpdateInstallment(_ context.Context, _ *model.LoanInstallment) error {
	return f.instErr
}

func (f *fakeLoanRepo) FindActiveVariableRateLoans(_ context.Context) ([]model.Loan, error) {
	return f.loans, nil
}

// ── Fake Loan Type Repository ────────────────────────────────────────────────

type fakeLoanTypeRepo struct {
	loanType *model.LoanType
	findErr  error
}

func (f *fakeLoanTypeRepo) FindByID(_ context.Context, _ uint) (*model.LoanType, error) {
	return f.loanType, f.findErr
}

// ── Fake Account Repository for Loan Tests ───────────────────────────────────

type fakeLoanAccountRepo struct {
	account  *model.Account
	accounts map[string]*model.Account
	findErr  error
}

func (f *fakeLoanAccountRepo) Create(_ context.Context, _ *model.Account) error { return nil }
func (f *fakeLoanAccountRepo) AccountNumberExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (f *fakeLoanAccountRepo) FindByAccountNumber(_ context.Context, accountNumber string) (*model.Account, error) {
	if f.accounts != nil {
		return f.accounts[accountNumber], f.findErr
	}
	return f.account, f.findErr
}
func (f *fakeLoanAccountRepo) GetByAccountNumber(_ context.Context, _ string) (*model.Account, error) {
	return f.account, f.findErr
}
func (f *fakeLoanAccountRepo) Update(_ context.Context, _ *model.Account) error { return nil }
func (f *fakeLoanAccountRepo) FindAllByClientID(_ context.Context, _ uint) ([]model.Account, error) {
	return nil, nil
}
func (f *fakeLoanAccountRepo) FindByAccountNumberAndClientID(_ context.Context, _ string, _ uint) (*model.Account, error) {
	return nil, nil
}
func (f *fakeLoanAccountRepo) NameExistsForClient(_ context.Context, _ uint, _ string, _ string) (bool, error) {
	return false, nil
}
func (f *fakeLoanAccountRepo) UpdateName(_ context.Context, _ string, _ string) error { return nil }
func (f *fakeLoanAccountRepo) UpdateLimits(_ context.Context, _ string, _ float64, _ float64) error {
	return nil
}
func (f *fakeLoanAccountRepo) UpdateBalance(_ context.Context, _ *model.Account) error { return nil }
func (f *fakeLoanAccountRepo) FindAll(_ context.Context, _ *dto.ListAccountsQuery) ([]*model.Account, int64, error) {
	return nil, 0, nil
}

// ── Fake Transaction Repository ──────────────────────────────────────────────

type fakeLoanTransactionRepo struct {
	transaction *model.Transaction
	createErr   error
	findErr     error
	updateErr   error
}

func (f *fakeLoanTransactionRepo) Create(_ context.Context, transaction *model.Transaction) error {
	if f.createErr != nil {
		return f.createErr
	}
	transaction.TransactionID = 1
	f.transaction = new(*transaction)
	return nil
}

func (f *fakeLoanTransactionRepo) GetByID(_ context.Context, _ uint) (*model.Transaction, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.transaction, nil
}

func (f *fakeLoanTransactionRepo) GetByPayerAccountNumber(_ context.Context, _ string) ([]*model.Transaction, error) {
	return nil, nil
}

func (f *fakeLoanTransactionRepo) GetByRecipientAccountNumber(_ context.Context, _ string) ([]*model.Transaction, error) {
	return nil, nil
}

func (f *fakeLoanTransactionRepo) Update(_ context.Context, transaction *model.Transaction) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.transaction = new(*transaction)
	return nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

func newLoanService(
	accountRepo repository.AccountRepository,
	loanTypeRepo repository.LoanTypeRepository,
	loanRequestRepo repository.LoanRequestRepository,
	loanRepo repository.LoanRepository,
	userClient client.UserClient,
	mailer Mailer,
) *LoanService {
	if accountRepo == nil {
		accountRepo = &fakeLoanAccountRepo{
			accounts: map[string]*model.Account{
				"client-account": {
					AccountNumber:    "client-account",
					AvailableBalance: 1_000_000,
					DailyLimit:       10_000_000,
					MonthlyLimit:     100_000_000,
					Currency:         model.Currency{Code: model.RSD},
				},
				BankAccounts[model.RSD]: {
					AccountNumber:    BankAccounts[model.RSD],
					AvailableBalance: 1_000_000,
					DailyLimit:       10_000_000,
					MonthlyLimit:     100_000_000,
					Currency:         model.Currency{Code: model.RSD},
				},
			},
		}
	}

	if loanRequestRepo == nil {
		loanRequestRepo = &fakeLoanRequestRepo{}
	}

	txRepo := &fakeLoanTransactionRepo{}
	txProcessor := NewTransactionProcessor(accountRepo, txRepo, &fakeBankingTxManager{})
	return NewLoanService(accountRepo, loanTypeRepo, loanRequestRepo, loanRepo, txProcessor, &fakeBankingTxManager{}, userClient, mailer)
}

func testLoanType() *model.LoanType {
	return &model.LoanType{
		LoanTypeID:         1,
		Name:               "Cash Loan",
		BaseInterestRate:   3.0,
		BankMargin:         2.5,
		MinRepaymentPeriod: 6,
		MaxRepaymentPeriod: 60,
	}
}

func loanTestAccount() *model.Account {
	return &model.Account{
		AccountNumber: "4440001100000001",
		ClientID:      1,
		Currency:      model.Currency{Code: model.RSD},
	}
}

// ── CalculateMonthlyInstallment Tests ────────────────────────────────────────

func TestCalculateMonthlyInstallment(t *testing.T) {
	t.Parallel()

	svc := newLoanService(nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name     string
		amount   float64
		rate     float64
		months   int
		expected float64
	}{
		{name: "zero rate divides evenly", amount: 12000, rate: 0, months: 12, expected: 1000},
		{name: "zero rate and zero months returns zero", amount: 12000, rate: 0, months: 0, expected: 0},
		{name: "standard interest rate calculation", amount: 100000, rate: 5.5, months: 24, expected: 4409.57},
		{name: "single month with interest", amount: 10000, rate: 12, months: 1, expected: 10100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CalculateMonthlyInstallment(tt.amount, tt.rate, tt.months)
			require.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

// ── SubmitLoanRequest Tests ──────────────────────────────────────────────────

func TestSubmitLoanRequest(t *testing.T) {
	t.Parallel()

	lt := testLoanType()

	tests := []struct {
		name            string
		accountRepo     *fakeLoanAccountRepo
		loanTypeRepo    *fakeLoanTypeRepo
		loanRepo        *fakeLoanRepo
		loanRequestRepo *fakeLoanRequestRepo
		userClient      *fakeUserClient
		mailer          *fakeMailer
		req             *dto.CreateLoanRequest
		expectErr       bool
		errMsg          string
	}{
		{
			name:            "success",
			accountRepo:     &fakeLoanAccountRepo{account: loanTestAccount()},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: lt},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "4440001100000001",
				LoanTypeID:      1,
				Amount:          100000,
				RepaymentPeriod: 24,
			},
		},
		{
			name:            "account not found",
			accountRepo:     &fakeLoanAccountRepo{account: nil},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: lt},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "nonexistent",
				LoanTypeID:      1,
				Amount:          100000,
				RepaymentPeriod: 24,
			},
			expectErr: true,
			errMsg:    "account not found",
		},
		{
			name:            "account repo error",
			accountRepo:     &fakeLoanAccountRepo{findErr: fmt.Errorf("db error")},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: lt},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "4440001100000001",
				LoanTypeID:      1,
				Amount:          100000,
				RepaymentPeriod: 24,
			},
			expectErr: true,
		},
		{
			name:            "loan type not found",
			accountRepo:     &fakeLoanAccountRepo{account: loanTestAccount()},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: nil},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "4440001100000001",
				LoanTypeID:      999,
				Amount:          100000,
				RepaymentPeriod: 24,
			},
			expectErr: true,
			errMsg:    "credit type not found",
		},
		{
			name:            "repayment period below minimum",
			accountRepo:     &fakeLoanAccountRepo{account: loanTestAccount()},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: lt},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "4440001100000001",
				LoanTypeID:      1,
				Amount:          100000,
				RepaymentPeriod: 3,
			},
			expectErr: true,
			errMsg:    "repayment period is not valid",
		},
		{
			name:            "repayment period above maximum",
			accountRepo:     &fakeLoanAccountRepo{account: loanTestAccount()},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: lt},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "4440001100000001",
				LoanTypeID:      1,
				Amount:          100000,
				RepaymentPeriod: 120,
			},
			expectErr: true,
			errMsg:    "repayment period is not valid",
		},
		{
			name:            "repo create fails",
			accountRepo:     &fakeLoanAccountRepo{account: loanTestAccount()},
			loanTypeRepo:    &fakeLoanTypeRepo{loanType: lt},
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{createErr: fmt.Errorf("db error")},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			req: &dto.CreateLoanRequest{
				AccountNumber:   "4440001100000001",
				LoanTypeID:      1,
				Amount:          100000,
				RepaymentPeriod: 24,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newLoanService(tt.accountRepo, tt.loanTypeRepo, tt.loanRequestRepo, tt.loanRepo, tt.userClient, tt.mailer)

			resp, err := svc.SubmitLoanRequest(context.Background(), tt.req, 1)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Equal(t, model.LoanRequestPending, resp.Status)
			require.Greater(t, resp.MonthlyInstallment, 0.0)
		})
	}
}

// ── ApproveLoanRequest Tests ─────────────────────────────────────────────────

func TestApproveLoanRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		loanRepo        *fakeLoanRepo
		loanRequestRepo *fakeLoanRequestRepo
		userClient      *fakeUserClient
		mailer          *fakeMailer
		id              uint
		expectErr       bool
		errMsg          string
	}{
		{
			name:     "success",
			loanRepo: &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{
				request: &model.LoanRequest{
					ID:                 1,
					Status:             model.LoanRequestPending,
					AccountNumber:      "client-account",
					Amount:             100000,
					CalculatedRate:     5.5,
					MonthlyInstallment: 4409.57,
					RepaymentPeriod:    24,
				},
			},
			userClient: &fakeUserClient{},
			mailer:     &fakeMailer{},
			id:         1,
		},
		{
			name:            "request not found",
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{request: nil},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			id:              99,
			expectErr:       true,
			errMsg:          "loan request not found",
		},
		{
			name:     "already approved",
			loanRepo: &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{
				request: &model.LoanRequest{ID: 1, Status: model.LoanRequestApproved},
			},
			userClient: &fakeUserClient{},
			mailer:     &fakeMailer{},
			id:         1,
			expectErr:  true,
			errMsg:     "loan request is not pending",
		},
		{
			name:            "repo find error",
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{findErr: fmt.Errorf("db error")},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			id:              1,
			expectErr:       true,
		},
		{
			name:     "repo update error",
			loanRepo: &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{
				request:   &model.LoanRequest{ID: 1, Status: model.LoanRequestPending},
				updateErr: fmt.Errorf("db error"),
			},
			userClient: &fakeUserClient{},
			mailer:     &fakeMailer{},
			id:         1,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newLoanService(nil, nil, tt.loanRequestRepo, tt.loanRepo, tt.userClient, tt.mailer)

			err := svc.ApproveLoanRequest(context.Background(), tt.id)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, model.LoanRequestApproved, tt.loanRequestRepo.updated.Status)
		})
	}
}

// ── RejectLoanRequest Tests ──────────────────────────────────────────────────

func TestRejectLoanRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		loanRepo        *fakeLoanRepo
		loanRequestRepo *fakeLoanRequestRepo
		userClient      *fakeUserClient
		mailer          *fakeMailer
		id              uint
		expectErr       bool
		errMsg          string
	}{
		{
			name:     "success",
			loanRepo: &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{
				request: &model.LoanRequest{ID: 1, Status: model.LoanRequestPending},
			},
			userClient: &fakeUserClient{},
			mailer:     &fakeMailer{},
			id:         1,
		},
		{
			name:            "request not found",
			loanRepo:        &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{request: nil},
			userClient:      &fakeUserClient{},
			mailer:          &fakeMailer{},
			id:              99,
			expectErr:       true,
			errMsg:          "loan request not found",
		},
		{
			name:     "already rejected",
			loanRepo: &fakeLoanRepo{},
			loanRequestRepo: &fakeLoanRequestRepo{
				request: &model.LoanRequest{ID: 1, Status: model.LoanRequestRejected},
			},
			userClient: &fakeUserClient{},
			mailer:     &fakeMailer{},
			id:         1,
			expectErr:  true,
			errMsg:     "loan request is not pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newLoanService(nil, nil, tt.loanRequestRepo, tt.loanRepo, tt.userClient, tt.mailer)

			err := svc.RejectLoanRequest(context.Background(), tt.id)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			require.Equal(t, model.LoanRequestRejected, tt.loanRequestRepo.updated.Status)
		})
	}
}

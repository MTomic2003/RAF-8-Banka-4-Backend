package service

import (
	"context"
	"errors"
	"testing"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type fakeInterbankAccountRepo struct {
	accounts       map[string]*model.Account
	accountsByUser map[uint][]model.Account
	findErr        error
	updateErr      error
}

func newFakeInterbankAccountRepo(accounts ...*model.Account) *fakeInterbankAccountRepo {
	repo := &fakeInterbankAccountRepo{
		accounts:       map[string]*model.Account{},
		accountsByUser: map[uint][]model.Account{},
	}
	for _, account := range accounts {
		repo.accounts[account.AccountNumber] = account
		repo.accountsByUser[account.ClientID] = append(repo.accountsByUser[account.ClientID], *account)
	}
	return repo
}

func (r *fakeInterbankAccountRepo) Create(context.Context, *model.Account) error { return nil }
func (r *fakeInterbankAccountRepo) AccountNumberExists(context.Context, string) (bool, error) {
	return false, nil
}
func (r *fakeInterbankAccountRepo) GetByAccountNumber(ctx context.Context, accountNumber string) (*model.Account, error) {
	return r.FindByAccountNumber(ctx, accountNumber)
}
func (r *fakeInterbankAccountRepo) Update(context.Context, *model.Account) error { return nil }
func (r *fakeInterbankAccountRepo) FindAllByClientID(context.Context, uint) ([]model.Account, error) {
	return nil, nil
}
func (r *fakeInterbankAccountRepo) FindByAccountNumberAndClientID(context.Context, string, uint) (*model.Account, error) {
	return nil, nil
}
func (r *fakeInterbankAccountRepo) UpdateName(context.Context, string, string) error { return nil }
func (r *fakeInterbankAccountRepo) UpdateLimits(context.Context, string, float64, float64) error {
	return nil
}
func (r *fakeInterbankAccountRepo) NameExistsForClient(context.Context, uint, string, string) (bool, error) {
	return false, nil
}
func (r *fakeInterbankAccountRepo) FindByAccountNumber(_ context.Context, accountNumber string) (*model.Account, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.accounts[accountNumber], nil
}
func (r *fakeInterbankAccountRepo) UpdateBalance(_ context.Context, account *model.Account) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	r.accounts[account.AccountNumber] = account
	return nil
}
func (r *fakeInterbankAccountRepo) FindAll(context.Context, *dto.ListAccountsQuery) ([]*model.Account, int64, error) {
	return nil, 0, nil
}
func (r *fakeInterbankAccountRepo) FindByClientID(_ context.Context, clientID uint) ([]model.Account, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.accountsByUser[clientID], nil
}
func (r *fakeInterbankAccountRepo) FindByAccountType(context.Context, model.AccountType) (*model.Account, error) {
	return nil, nil
}

type fakeInterbankPostingRepo struct {
	rows    map[string]*model.InterbankCashPosting
	findErr error
	saveErr error
}

func newFakeInterbankPostingRepo(rows ...*model.InterbankCashPosting) *fakeInterbankPostingRepo {
	repo := &fakeInterbankPostingRepo{rows: map[string]*model.InterbankCashPosting{}}
	for _, row := range rows {
		cp := *row
		repo.rows[row.PostingID] = &cp
	}
	return repo
}

func (r *fakeInterbankPostingRepo) Create(_ context.Context, posting *model.InterbankCashPosting) error {
	cp := *posting
	r.rows[posting.PostingID] = &cp
	return nil
}

func (r *fakeInterbankPostingRepo) FindByID(_ context.Context, postingID string) (*model.InterbankCashPosting, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	row := r.rows[postingID]
	if row == nil {
		return nil, nil
	}
	cp := *row
	return &cp, nil
}

func (r *fakeInterbankPostingRepo) Save(_ context.Context, posting *model.InterbankCashPosting) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	cp := *posting
	r.rows[posting.PostingID] = &cp
	return nil
}

type fakeInterbankConverter struct {
	result float64
	err    error
}

func (c fakeInterbankConverter) Convert(context.Context, float64, model.CurrencyCode, model.CurrencyCode) (float64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return c.result, nil
}

func (c fakeInterbankConverter) CalculateFee(amount float64) float64 { return amount }

func interbankAccount(number string, clientID uint, code model.CurrencyCode, balance, available float64) *model.Account {
	return &model.Account{
		AccountNumber:    number,
		ClientID:         clientID,
		Currency:         model.Currency{Code: code},
		Balance:          balance,
		AvailableBalance: available,
		Status:           "Active",
	}
}

func TestInterbankCashPrepareExplicitAccountAndIdempotency(t *testing.T) {
	t.Parallel()

	account := interbankAccount("444000000000000011", 1, model.RSD, 1000, 1000)
	accounts := newFakeInterbankAccountRepo(account)
	postings := newFakeInterbankPostingRepo()
	svc := NewInterbankCashService(accounts, postings, &fakeBankingTxManager{}, fakeInterbankConverter{})

	posting, err := svc.Prepare(context.Background(), "posting-1", account.AccountNumber, 0, model.RSD, -250, "CLIENT")
	if err != nil {
		t.Fatalf("prepare explicit account: %v", err)
	}
	if posting.Status != model.InterbankCashPostingPrepared || posting.AccountNumber != account.AccountNumber {
		t.Fatalf("unexpected posting %#v", posting)
	}
	if accounts.accounts[account.AccountNumber].AvailableBalance != 750 {
		t.Fatalf("available balance = %.2f, want 750", accounts.accounts[account.AccountNumber].AvailableBalance)
	}

	again, err := svc.Prepare(context.Background(), "posting-1", account.AccountNumber, 0, model.RSD, -250, "CLIENT")
	if err != nil {
		t.Fatalf("idempotent prepare: %v", err)
	}
	if again.PostingID != "posting-1" || accounts.accounts[account.AccountNumber].AvailableBalance != 750 {
		t.Fatalf("unexpected idempotent prepare result %#v", again)
	}

	_, err = svc.Prepare(context.Background(), "posting-1", account.AccountNumber, 0, model.RSD, -251, "CLIENT")
	if err == nil {
		t.Fatal("expected conflict for reused posting id with different amount")
	}
}

func TestInterbankCashPrepareChoosesClientAccountAndConverts(t *testing.T) {
	t.Parallel()

	eur := interbankAccount("444000000000000021", 7, model.EUR, 500, 500)
	rsd := interbankAccount("444000000000000022", 7, model.RSD, 20000, 20000)
	accounts := newFakeInterbankAccountRepo(eur, rsd)
	postings := newFakeInterbankPostingRepo()
	svc := NewInterbankCashService(accounts, postings, &fakeBankingTxManager{}, fakeInterbankConverter{result: -11700})

	posting, err := svc.Prepare(context.Background(), "posting-2", "", 7, model.USD, -100, "CLIENT")
	if err != nil {
		t.Fatalf("prepare client account: %v", err)
	}
	if posting.AccountNumber != rsd.AccountNumber || posting.CurrencyCode != model.RSD || posting.Amount != -11700 {
		t.Fatalf("unexpected converted posting %#v", posting)
	}
	if accounts.accounts[rsd.AccountNumber].AvailableBalance != 8300 {
		t.Fatalf("available balance = %.2f, want 8300 after converted reserve", accounts.accounts[rsd.AccountNumber].AvailableBalance)
	}
}

func TestInterbankCashPrepareEmployeeUsesBankAccount(t *testing.T) {
	t.Parallel()

	bankAccount := interbankAccount(BankAccounts[model.RSD], 0, model.RSD, 100000, 100000)
	accounts := newFakeInterbankAccountRepo(bankAccount)
	svc := NewInterbankCashService(accounts, newFakeInterbankPostingRepo(), &fakeBankingTxManager{}, fakeInterbankConverter{})

	posting, err := svc.Prepare(context.Background(), "posting-employee", "", 0, model.RSD, 500, "EMPLOYEE")
	if err != nil {
		t.Fatalf("prepare employee posting: %v", err)
	}
	if posting.AccountNumber != BankAccounts[model.RSD] || posting.Amount != 500 {
		t.Fatalf("unexpected employee posting %#v", posting)
	}
}

func TestInterbankCashCommitAndRollbackTransitions(t *testing.T) {
	t.Parallel()

	account := interbankAccount("444000000000000031", 1, model.RSD, 1000, 800)
	posting := &model.InterbankCashPosting{
		PostingID:             "posting-commit",
		AccountNumber:         account.AccountNumber,
		CurrencyCode:          model.RSD,
		Amount:                -200,
		RequestedCurrencyCode: model.RSD,
		RequestedAmount:       -200,
		Status:                model.InterbankCashPostingPrepared,
	}
	accounts := newFakeInterbankAccountRepo(account)
	postings := newFakeInterbankPostingRepo(posting)
	svc := NewInterbankCashService(accounts, postings, &fakeBankingTxManager{}, fakeInterbankConverter{})

	committed, err := svc.Commit(context.Background(), "posting-commit")
	if err != nil {
		t.Fatalf("commit posting: %v", err)
	}
	if committed.Status != model.InterbankCashPostingCommitted || accounts.accounts[account.AccountNumber].Balance != 800 {
		t.Fatalf("unexpected committed state posting=%#v account=%#v", committed, accounts.accounts[account.AccountNumber])
	}

	again, err := svc.Commit(context.Background(), "posting-commit")
	if err != nil {
		t.Fatalf("idempotent commit: %v", err)
	}
	if again.Status != model.InterbankCashPostingCommitted || accounts.accounts[account.AccountNumber].Balance != 800 {
		t.Fatalf("unexpected second commit state posting=%#v account=%#v", again, accounts.accounts[account.AccountNumber])
	}

	creditAccount := interbankAccount("444000000000000032", 2, model.RSD, 1000, 1000)
	creditPosting := &model.InterbankCashPosting{
		PostingID:             "posting-credit",
		AccountNumber:         creditAccount.AccountNumber,
		CurrencyCode:          model.RSD,
		Amount:                300,
		RequestedCurrencyCode: model.RSD,
		RequestedAmount:       300,
		Status:                model.InterbankCashPostingPrepared,
	}
	accounts.accounts[creditAccount.AccountNumber] = creditAccount
	postings.rows[creditPosting.PostingID] = creditPosting
	credited, err := svc.Commit(context.Background(), "posting-credit")
	if err != nil {
		t.Fatalf("commit credit posting: %v", err)
	}
	if credited.Status != model.InterbankCashPostingCommitted || accounts.accounts[creditAccount.AccountNumber].Balance != 1300 || accounts.accounts[creditAccount.AccountNumber].AvailableBalance != 1300 {
		t.Fatalf("unexpected credited state posting=%#v account=%#v", credited, accounts.accounts[creditAccount.AccountNumber])
	}
}

func TestInterbankCashRollbackAndErrors(t *testing.T) {
	t.Parallel()

	account := interbankAccount("444000000000000041", 1, model.RSD, 1000, 700)
	prepared := &model.InterbankCashPosting{
		PostingID:             "posting-rollback",
		AccountNumber:         account.AccountNumber,
		CurrencyCode:          model.RSD,
		Amount:                -300,
		RequestedCurrencyCode: model.RSD,
		RequestedAmount:       -300,
		Status:                model.InterbankCashPostingPrepared,
	}
	committed := &model.InterbankCashPosting{
		PostingID:             "posting-committed",
		AccountNumber:         account.AccountNumber,
		CurrencyCode:          model.RSD,
		Amount:                -100,
		RequestedCurrencyCode: model.RSD,
		RequestedAmount:       -100,
		Status:                model.InterbankCashPostingCommitted,
	}
	accounts := newFakeInterbankAccountRepo(account)
	postings := newFakeInterbankPostingRepo(prepared, committed)
	svc := NewInterbankCashService(accounts, postings, &fakeBankingTxManager{}, fakeInterbankConverter{})

	rolledBack, err := svc.Rollback(context.Background(), "posting-rollback")
	if err != nil {
		t.Fatalf("rollback posting: %v", err)
	}
	if rolledBack.Status != model.InterbankCashPostingRolledBack || accounts.accounts[account.AccountNumber].AvailableBalance != 1000 {
		t.Fatalf("unexpected rollback state posting=%#v account=%#v", rolledBack, accounts.accounts[account.AccountNumber])
	}

	again, err := svc.Rollback(context.Background(), "posting-rollback")
	if err != nil {
		t.Fatalf("idempotent rollback: %v", err)
	}
	if again.Status != model.InterbankCashPostingRolledBack || accounts.accounts[account.AccountNumber].AvailableBalance != 1000 {
		t.Fatalf("unexpected second rollback state posting=%#v account=%#v", again, accounts.accounts[account.AccountNumber])
	}

	if _, err := svc.Rollback(context.Background(), "posting-committed"); err == nil {
		t.Fatal("expected error when rolling back committed posting")
	}
	if _, err := svc.Commit(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for missing posting")
	}
	if _, err := svc.Prepare(context.Background(), "", account.AccountNumber, 0, model.RSD, -1, "CLIENT"); err == nil {
		t.Fatal("expected error for empty posting id")
	}
	if _, err := svc.Prepare(context.Background(), "bad-zero", account.AccountNumber, 0, model.RSD, 0, "CLIENT"); err == nil {
		t.Fatal("expected error for zero amount")
	}
	if _, err := svc.Prepare(context.Background(), "bad-currency", account.AccountNumber, 0, model.CurrencyCode("BAD"), -1, "CLIENT"); err == nil {
		t.Fatal("expected error for unsupported currency")
	}

	converterErr := NewInterbankCashService(accounts, newFakeInterbankPostingRepo(), &fakeBankingTxManager{}, fakeInterbankConverter{err: errors.New("rates unavailable")})
	if _, err := converterErr.Prepare(context.Background(), "bad-convert", "", 1, model.USD, -1, "CLIENT"); err == nil {
		t.Fatal("expected converter error")
	}
}

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

// ── Fakes scoped to interbank cash posting record-keeping ──────────────────

type fakeCashPostingRepo struct {
	postings  map[string]*model.InterbankCashPosting
	saveCount int
}

func newFakeCashPostingRepo() *fakeCashPostingRepo {
	return &fakeCashPostingRepo{postings: map[string]*model.InterbankCashPosting{}}
}

func (f *fakeCashPostingRepo) Create(_ context.Context, p *model.InterbankCashPosting) error {
	f.postings[p.PostingID] = p
	return nil
}

func (f *fakeCashPostingRepo) FindByID(_ context.Context, id string) (*model.InterbankCashPosting, error) {
	return f.postings[id], nil
}

func (f *fakeCashPostingRepo) Save(_ context.Context, p *model.InterbankCashPosting) error {
	f.postings[p.PostingID] = p
	f.saveCount++
	return nil
}

// recTxRepo / recPayRepo count created records so tests can assert exactly-once.

type recTxRepo struct {
	created  []*model.Transaction
	existing *model.Transaction // returned by GetByID for the dedup gate
}

func (r *recTxRepo) Create(_ context.Context, t *model.Transaction) error {
	t.TransactionID = uint(len(r.created) + 1)
	r.created = append(r.created, t)
	return nil
}
func (r *recTxRepo) Update(_ context.Context, _ *model.Transaction) error { return nil }
func (r *recTxRepo) GetByID(_ context.Context, _ uint) (*model.Transaction, error) {
	return r.existing, nil
}
func (r *recTxRepo) GetByPayerAccountNumber(_ context.Context, _ string) ([]*model.Transaction, error) {
	return nil, nil
}
func (r *recTxRepo) GetByRecipientAccountNumber(_ context.Context, _ string) ([]*model.Transaction, error) {
	return nil, nil
}

type recPayRepo struct {
	created []*model.Payment
}

func (r *recPayRepo) Create(_ context.Context, p *model.Payment) error {
	r.created = append(r.created, p)
	return nil
}
func (r *recPayRepo) GetByID(_ context.Context, _ uint) (*model.Payment, error) { return nil, nil }
func (r *recPayRepo) Update(_ context.Context, _ *model.Payment) error          { return nil }
func (r *recPayRepo) FindByAccount(_ context.Context, _ string, _ *dto.PaymentFilters) ([]model.Payment, int64, error) {
	return nil, 0, nil
}
func (r *recPayRepo) FindByClient(_ context.Context, _ uint, _ *dto.PaymentFilters) ([]model.Payment, int64, error) {
	return nil, 0, nil
}

func newCashService(acc *fakePaymentAccountRepo, posts *fakeCashPostingRepo, txRepo *recTxRepo, payRepo *recPayRepo) *InterbankCashService {
	return NewInterbankCashService(acc, posts, &fakeBankingTxManager{}, &fakeCurrencyConverter{}, txRepo, payRepo)
}

func localAccount(number string, balance float64) *model.Account {
	return &model.Account{
		AccountNumber:    number,
		Balance:          balance,
		AvailableBalance: balance,
		Currency:         model.Currency{Code: model.RSD},
	}
}

// ── Commit creates Transaction+Payment history records ─────────────────────

func TestInterbankCashService_Commit_CreatesHistory(t *testing.T) {
	const local = "444000000000000011"
	const counterparty = "111000000000000022"

	cases := []struct {
		name          string
		amount        float64 // resolved amount; sign drives payer/recipient
		bankingTxID   uint64
		existingTx    *model.Transaction
		wantRecords   bool
		wantPayer     string
		wantRecipient string
	}{
		{
			name:          "incoming credit records local as recipient",
			amount:        100,
			wantRecords:   true,
			wantPayer:     counterparty,
			wantRecipient: local,
		},
		{
			name:          "outgoing debit records local as payer",
			amount:        -100,
			wantRecords:   true,
			wantPayer:     local,
			wantRecipient: counterparty,
		},
		{
			name:        "initiating payment leg with existing transaction is skipped",
			amount:      -100,
			bankingTxID: 7,
			existingTx:  &model.Transaction{TransactionID: 7},
			wantRecords: false,
		},
		{
			name:          "initiating leg whose transaction is missing still records",
			amount:        -100,
			bankingTxID:   7,
			existingTx:    nil,
			wantRecords:   true,
			wantPayer:     local,
			wantRecipient: counterparty,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			acc := newFakePaymentAccountRepo(localAccount(local, 1000))
			posts := newFakeCashPostingRepo()
			txRepo := &recTxRepo{existing: tc.existingTx}
			payRepo := &recPayRepo{}
			svc := newCashService(acc, posts, txRepo, payRepo)

			posts.postings["pid-1"] = &model.InterbankCashPosting{
				PostingID:                 "pid-1",
				AccountNumber:             local,
				CurrencyCode:              model.RSD,
				Amount:                    tc.amount,
				RequestedCurrencyCode:     model.RSD,
				RequestedAmount:           tc.amount,
				Status:                    model.InterbankCashPostingPrepared,
				BankingTxID:               tc.bankingTxID,
				CounterpartyAccountNumber: counterparty,
				PaymentCode:               "289",
				Purpose:                   "interbank transfer",
			}

			_, err := svc.Commit(context.Background(), "pid-1")
			require.NoError(t, err)

			if !tc.wantRecords {
				require.Empty(t, txRepo.created, "expected no transaction record")
				require.Empty(t, payRepo.created, "expected no payment record")
				return
			}

			require.Len(t, txRepo.created, 1)
			require.Len(t, payRepo.created, 1)
			rec := txRepo.created[0]
			require.Equal(t, model.TransactionCompleted, rec.Status)
			require.Equal(t, 100.0, rec.StartAmount)
			require.Equal(t, 100.0, rec.EndAmount)
			require.Equal(t, tc.wantPayer, rec.PayerAccountNumber)
			require.Equal(t, tc.wantRecipient, rec.RecipientAccountNumber)
			require.Equal(t, "289", payRepo.created[0].PaymentCode)
			require.Equal(t, "interbank transfer", payRepo.created[0].Purpose)
			require.Equal(t, rec.TransactionID, payRepo.created[0].TransactionID)
		})
	}
}

func TestInterbankCashService_Commit_Idempotent(t *testing.T) {
	const local = "444000000000000011"
	acc := newFakePaymentAccountRepo(localAccount(local, 1000))
	posts := newFakeCashPostingRepo()
	txRepo := &recTxRepo{}
	payRepo := &recPayRepo{}
	svc := newCashService(acc, posts, txRepo, payRepo)

	posts.postings["pid-1"] = &model.InterbankCashPosting{
		PostingID:             "pid-1",
		AccountNumber:         local,
		CurrencyCode:          model.RSD,
		Amount:                100,
		RequestedCurrencyCode: model.RSD,
		RequestedAmount:       100,
		Status:                model.InterbankCashPostingPrepared,
	}

	_, err := svc.Commit(context.Background(), "pid-1")
	require.NoError(t, err)
	// A retransmitted COMMIT_TX must not create a second record.
	_, err = svc.Commit(context.Background(), "pid-1")
	require.NoError(t, err)

	require.Len(t, txRepo.created, 1, "record must be created exactly once")
	require.Len(t, payRepo.created, 1)
}

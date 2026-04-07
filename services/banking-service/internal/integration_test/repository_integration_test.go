//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoAccount_GetByAccountNumber(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAccountRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 10, curr.CurrencyID, 3000)

	found, err := repo.GetByAccountNumber(context.Background(), acc.AccountNumber)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, acc.AccountNumber, found.AccountNumber)
}

func TestRepoAccount_GetByAccountNumber_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAccountRepository(db)

	found, err := repo.GetByAccountNumber(context.Background(), "000000000000000000")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRepoAccount_Update(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAccountRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 20, curr.CurrencyID, 1000)

	acc.Name = "Updated Account Name"
	acc.Balance = 9999
	err := repo.Update(context.Background(), acc)
	require.NoError(t, err)

	updated, err := repo.GetByAccountNumber(context.Background(), acc.AccountNumber)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Account Name", updated.Name)
	assert.Equal(t, 9999.0, updated.Balance)
}

func TestRepoAccount_FindByClientID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAccountRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	clientID := uint(30)
	acc1 := seedAccount(t, db, clientID, curr.CurrencyID, 500)
	acc2 := seedAccount(t, db, clientID, curr.CurrencyID, 750)
	// different client — should not appear
	seedAccount(t, db, 999, curr.CurrencyID, 100)

	accounts, err := repo.FindByClientID(context.Background(), clientID)
	require.NoError(t, err)
	require.Len(t, accounts, 2)

	numbers := []string{accounts[0].AccountNumber, accounts[1].AccountNumber}
	assert.Contains(t, numbers, acc1.AccountNumber)
	assert.Contains(t, numbers, acc2.AccountNumber)
}

func TestRepoAccount_FindByClientID_Empty(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAccountRepository(db)

	accounts, err := repo.FindByClientID(context.Background(), 88888)
	require.NoError(t, err)
	assert.Empty(t, accounts)
}

func TestRepoLoan_FindLoanByRequestID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 40, curr.CurrencyID, 0)
	lt := seedLoanType(t, db)

	req := &model.LoanRequest{
		ClientID:        40,
		AccountNumber:   acc.AccountNumber,
		LoanTypeID:      lt.LoanTypeID,
		Amount:          50000,
		RepaymentPeriod: 24,
		Status:          model.LoanRequestApproved,
	}
	require.NoError(t, db.Create(req).Error)

	loan := &model.Loan{
		LoanRequestID:       req.ID,
		MonthlyInstallment:  2300,
		InterestRate:        5.5,
		IsVariableRate:      false,
		RemainingDebt:       50000,
		RepaymentPeriod:     24,
		StartDate:           time.Now(),
		NextInstallmentDate: time.Now().Add(30 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	require.NoError(t, db.Create(loan).Error)

	found, err := repo.FindLoanByRequestID(context.Background(), req.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, loan.ID, found.ID)
}

func TestRepoLoan_FindLoanByRequestID_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	found, err := repo.FindLoanByRequestID(context.Background(), 99999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRepoLoan_UpdateLoan(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 41, curr.CurrencyID, 0)
	lt := seedLoanType(t, db)

	req := &model.LoanRequest{
		ClientID:        41,
		AccountNumber:   acc.AccountNumber,
		LoanTypeID:      lt.LoanTypeID,
		Amount:          20000,
		RepaymentPeriod: 12,
		Status:          model.LoanRequestApproved,
	}
	require.NoError(t, db.Create(req).Error)

	loan := &model.Loan{
		LoanRequestID:       req.ID,
		MonthlyInstallment:  1800,
		InterestRate:        4.0,
		RemainingDebt:       20000,
		RepaymentPeriod:     12,
		StartDate:           time.Now(),
		NextInstallmentDate: time.Now().Add(30 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	require.NoError(t, db.Create(loan).Error)

	loan.RemainingDebt = 18000
	loan.PaidInstallments = 1
	loan.Status = model.LoanStatusActive
	err := repo.UpdateLoan(context.Background(), loan)
	require.NoError(t, err)

	var updated model.Loan
	require.NoError(t, db.First(&updated, loan.ID).Error)
	assert.Equal(t, 18000.0, updated.RemainingDebt)
	assert.Equal(t, 1, updated.PaidInstallments)
}

func TestRepoLoan_FindDueInstallments(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 42, curr.CurrencyID, 0)
	lt := seedLoanType(t, db)

	req := &model.LoanRequest{
		ClientID:        42,
		AccountNumber:   acc.AccountNumber,
		LoanTypeID:      lt.LoanTypeID,
		Amount:          30000,
		RepaymentPeriod: 12,
		Status:          model.LoanRequestApproved,
	}
	require.NoError(t, db.Create(req).Error)

	loan := &model.Loan{
		LoanRequestID:       req.ID,
		MonthlyInstallment:  2600,
		InterestRate:        5.0,
		RemainingDebt:       30000,
		RepaymentPeriod:     12,
		StartDate:           time.Now().Add(-60 * 24 * time.Hour),
		NextInstallmentDate: time.Now().Add(-1 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	require.NoError(t, db.Create(loan).Error)

	pastDue := &model.LoanInstallment{
		LoanID:            loan.ID,
		InstallmentNumber: 1,
		Amount:            2600,
		InterestRate:      5.0,
		DueDate:           time.Now().Add(-2 * 24 * time.Hour),
		Status:            model.InstallmentStatusPending,
	}
	futureDue := &model.LoanInstallment{
		LoanID:            loan.ID,
		InstallmentNumber: 2,
		Amount:            2600,
		InterestRate:      5.0,
		DueDate:           time.Now().Add(28 * 24 * time.Hour),
		Status:            model.InstallmentStatusPending,
	}
	require.NoError(t, db.Create(pastDue).Error)
	require.NoError(t, db.Create(futureDue).Error)

	due, err := repo.FindDueInstallments(context.Background(), time.Now())
	require.NoError(t, err)
	require.Len(t, due, 1)
	assert.Equal(t, pastDue.ID, due[0].ID)
}

func TestRepoLoan_FindRetryInstallments(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 43, curr.CurrencyID, 0)
	lt := seedLoanType(t, db)

	req := &model.LoanRequest{
		ClientID:        43,
		AccountNumber:   acc.AccountNumber,
		LoanTypeID:      lt.LoanTypeID,
		Amount:          15000,
		RepaymentPeriod: 6,
		Status:          model.LoanRequestApproved,
	}
	require.NoError(t, db.Create(req).Error)

	loan := &model.Loan{
		LoanRequestID:       req.ID,
		MonthlyInstallment:  2600,
		InterestRate:        5.0,
		RemainingDebt:       15000,
		RepaymentPeriod:     6,
		StartDate:           time.Now().Add(-30 * 24 * time.Hour),
		NextInstallmentDate: time.Now().Add(-1 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	require.NoError(t, db.Create(loan).Error)

	pastRetry := time.Now().Add(-1 * time.Hour)
	futureRetry := time.Now().Add(2 * time.Hour)

	readyInstallment := &model.LoanInstallment{
		LoanID:            loan.ID,
		InstallmentNumber: 1,
		Amount:            2600,
		InterestRate:      5.0,
		DueDate:           time.Now().Add(-30 * 24 * time.Hour),
		RetryAt:           &pastRetry,
		Status:            model.InstallmentStatusRetrying,
	}
	notReadyInstallment := &model.LoanInstallment{
		LoanID:            loan.ID,
		InstallmentNumber: 2,
		Amount:            2600,
		InterestRate:      5.0,
		DueDate:           time.Now().Add(-5 * 24 * time.Hour),
		RetryAt:           &futureRetry,
		Status:            model.InstallmentStatusRetrying,
	}
	require.NoError(t, db.Create(readyInstallment).Error)
	require.NoError(t, db.Create(notReadyInstallment).Error)

	retries, err := repo.FindRetryInstallments(context.Background(), time.Now())
	require.NoError(t, err)
	require.Len(t, retries, 1)
	assert.Equal(t, readyInstallment.ID, retries[0].ID)
}

func TestRepoLoan_UpdateInstallment(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 44, curr.CurrencyID, 0)
	lt := seedLoanType(t, db)

	req := &model.LoanRequest{
		ClientID:        44,
		AccountNumber:   acc.AccountNumber,
		LoanTypeID:      lt.LoanTypeID,
		Amount:          10000,
		RepaymentPeriod: 6,
		Status:          model.LoanRequestApproved,
	}
	require.NoError(t, db.Create(req).Error)

	loan := &model.Loan{
		LoanRequestID:       req.ID,
		MonthlyInstallment:  1700,
		InterestRate:        4.5,
		RemainingDebt:       10000,
		RepaymentPeriod:     6,
		StartDate:           time.Now(),
		NextInstallmentDate: time.Now().Add(30 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	require.NoError(t, db.Create(loan).Error)

	installment := &model.LoanInstallment{
		LoanID:            loan.ID,
		InstallmentNumber: 1,
		Amount:            1700,
		InterestRate:      4.5,
		DueDate:           time.Now().Add(30 * 24 * time.Hour),
		Status:            model.InstallmentStatusPending,
	}
	require.NoError(t, db.Create(installment).Error)

	now := time.Now()
	installment.Status = model.InstallmentStatusPaid
	installment.PaidAt = &now
	err := repo.UpdateInstallment(context.Background(), installment)
	require.NoError(t, err)

	var updated model.LoanInstallment
	require.NoError(t, db.First(&updated, installment.ID).Error)
	assert.Equal(t, model.InstallmentStatusPaid, updated.Status)
	assert.NotNil(t, updated.PaidAt)
}

func TestRepoLoan_FindActiveVariableRateLoans(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewLoanRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 45, curr.CurrencyID, 0)
	lt := seedLoanType(t, db)

	makeReq := func(clientID uint) *model.LoanRequest {
		r := &model.LoanRequest{
			ClientID:        clientID,
			AccountNumber:   acc.AccountNumber,
			LoanTypeID:      lt.LoanTypeID,
			Amount:          25000,
			RepaymentPeriod: 12,
			Status:          model.LoanRequestApproved,
		}
		require.NoError(t, db.Create(r).Error)
		return r
	}

	req1 := makeReq(45)
	req2 := makeReq(46)
	req3 := makeReq(47)

	variableActive := &model.Loan{
		LoanRequestID:       req1.ID,
		MonthlyInstallment:  2200,
		InterestRate:        3.5,
		IsVariableRate:      true,
		RemainingDebt:       25000,
		RepaymentPeriod:     12,
		StartDate:           time.Now(),
		NextInstallmentDate: time.Now().Add(30 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	fixedActive := &model.Loan{
		LoanRequestID:       req2.ID,
		MonthlyInstallment:  2200,
		InterestRate:        3.5,
		IsVariableRate:      false,
		RemainingDebt:       25000,
		RepaymentPeriod:     12,
		StartDate:           time.Now(),
		NextInstallmentDate: time.Now().Add(30 * 24 * time.Hour),
		Status:              model.LoanStatusActive,
	}
	variableCompleted := &model.Loan{
		LoanRequestID:       req3.ID,
		MonthlyInstallment:  2200,
		InterestRate:        3.5,
		IsVariableRate:      true,
		RemainingDebt:       0,
		RepaymentPeriod:     12,
		StartDate:           time.Now().Add(-400 * 24 * time.Hour),
		NextInstallmentDate: time.Now().Add(-10 * 24 * time.Hour),
		Status:              model.LoanStatusCompleted,
	}
	require.NoError(t, db.Create(variableActive).Error)
	require.NoError(t, db.Create(fixedActive).Error)
	require.NoError(t, db.Create(variableCompleted).Error)

	loans, err := repo.FindActiveVariableRateLoans(context.Background())
	require.NoError(t, err)
	require.Len(t, loans, 1)
	assert.Equal(t, variableActive.ID, loans[0].ID)
	assert.True(t, loans[0].IsVariableRate)
	assert.Equal(t, model.LoanStatusActive, loans[0].Status)
}

func TestRepoPayment_Update(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewPaymentRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	payer := seedAccount(t, db, 50, curr.CurrencyID, 5000)
	recipient := seedAccount(t, db, 51, curr.CurrencyID, 1000)

	tx := &model.Transaction{
		PayerAccountNumber:     payer.AccountNumber,
		RecipientAccountNumber: recipient.AccountNumber,
		StartAmount:            200,
		StartCurrencyCode:      model.RSD,
		EndAmount:              200,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionProcessing,
	}
	require.NoError(t, db.Create(tx).Error)

	payment := &model.Payment{
		TransactionID:   tx.TransactionID,
		RecipientName:   "Test Recipient",
		ReferenceNumber: "REF001",
		PaymentCode:     "289",
		Purpose:         "Test payment",
	}
	require.NoError(t, db.Create(payment).Error)

	payment.RecipientName = "Updated Recipient"
	payment.FailedAttempts = 2
	err := repo.Update(context.Background(), payment)
	require.NoError(t, err)

	var updated model.Payment
	require.NoError(t, db.First(&updated, payment.PaymentID).Error)
	assert.Equal(t, "Updated Recipient", updated.RecipientName)
	assert.Equal(t, 2, updated.FailedAttempts)
}

func TestRepoTransaction_GetByPayerAccountNumber(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewTransactionRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	payer := seedAccount(t, db, 60, curr.CurrencyID, 10000)
	recipient := seedAccount(t, db, 61, curr.CurrencyID, 500)
	other := seedAccount(t, db, 62, curr.CurrencyID, 500)

	tx1 := &model.Transaction{
		PayerAccountNumber:     payer.AccountNumber,
		RecipientAccountNumber: recipient.AccountNumber,
		StartAmount:            100,
		StartCurrencyCode:      model.RSD,
		EndAmount:              100,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionCompleted,
	}
	tx2 := &model.Transaction{
		PayerAccountNumber:     payer.AccountNumber,
		RecipientAccountNumber: other.AccountNumber,
		StartAmount:            200,
		StartCurrencyCode:      model.RSD,
		EndAmount:              200,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionCompleted,
	}
	// unrelated transaction
	txOther := &model.Transaction{
		PayerAccountNumber:     recipient.AccountNumber,
		RecipientAccountNumber: other.AccountNumber,
		StartAmount:            50,
		StartCurrencyCode:      model.RSD,
		EndAmount:              50,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionCompleted,
	}
	require.NoError(t, db.Create(tx1).Error)
	require.NoError(t, db.Create(tx2).Error)
	require.NoError(t, db.Create(txOther).Error)

	txns, err := repo.GetByPayerAccountNumber(context.Background(), payer.AccountNumber)
	require.NoError(t, err)
	require.Len(t, txns, 2)

	ids := []uint{txns[0].TransactionID, txns[1].TransactionID}
	assert.Contains(t, ids, tx1.TransactionID)
	assert.Contains(t, ids, tx2.TransactionID)
}

func TestRepoTransaction_GetByRecipientAccountNumber(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewTransactionRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	payer := seedAccount(t, db, 70, curr.CurrencyID, 10000)
	recipient := seedAccount(t, db, 71, curr.CurrencyID, 500)
	other := seedAccount(t, db, 72, curr.CurrencyID, 500)

	tx1 := &model.Transaction{
		PayerAccountNumber:     payer.AccountNumber,
		RecipientAccountNumber: recipient.AccountNumber,
		StartAmount:            300,
		StartCurrencyCode:      model.RSD,
		EndAmount:              300,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionCompleted,
	}
	tx2 := &model.Transaction{
		PayerAccountNumber:     other.AccountNumber,
		RecipientAccountNumber: recipient.AccountNumber,
		StartAmount:            150,
		StartCurrencyCode:      model.RSD,
		EndAmount:              150,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionProcessing,
	}
	// different recipient — should not appear
	txOther := &model.Transaction{
		PayerAccountNumber:     payer.AccountNumber,
		RecipientAccountNumber: other.AccountNumber,
		StartAmount:            75,
		StartCurrencyCode:      model.RSD,
		EndAmount:              75,
		EndCurrencyCode:        model.RSD,
		Status:                 model.TransactionCompleted,
	}
	require.NoError(t, db.Create(tx1).Error)
	require.NoError(t, db.Create(tx2).Error)
	require.NoError(t, db.Create(txOther).Error)

	txns, err := repo.GetByRecipientAccountNumber(context.Background(), recipient.AccountNumber)
	require.NoError(t, err)
	require.Len(t, txns, 2)

	ids := []uint{txns[0].TransactionID, txns[1].TransactionID}
	assert.Contains(t, ids, tx1.TransactionID)
	assert.Contains(t, ids, tx2.TransactionID)
}

func TestRepoCard_CountByAccountNumber(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewCardRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 80, curr.CurrencyID, 1000)

	seedCard(t, db, acc.AccountNumber)
	seedCard(t, db, acc.AccountNumber)

	count, err := repo.CountByAccountNumber(context.Background(), acc.AccountNumber)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRepoCard_CountByAccountNumber_Empty(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewCardRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 81, curr.CurrencyID, 1000)

	count, err := repo.CountByAccountNumber(context.Background(), acc.AccountNumber)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepoCard_CountByAccountNumberAndAuthorizedPersonID_Nil(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewCardRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 82, curr.CurrencyID, 1000)

	// card with no authorized person
	seedCard(t, db, acc.AccountNumber)

	// create a real authorized person so the FK constraint is satisfied
	ap := &model.AuthorizedPerson{
		AccountNumber: acc.AccountNumber,
		FirstName:     "AP",
		LastName:      "Test",
		Email:         fmt.Sprintf("ap-%d@example.com", uniqueCounter.Add(1)),
	}
	require.NoError(t, db.Create(ap).Error)

	// card linked to an authorized person — use seedCard-style 16-digit number
	card2 := &model.Card{
		CardNumber:         fmt.Sprintf("4532000001%06d", uniqueCounter.Add(1)),
		CardType:           model.CardTypeDebit,
		CardBrand:          model.CardBrandVisa,
		Name:               "Auth Card",
		AccountNumber:      acc.AccountNumber,
		CVV:                "321",
		Limit:              500,
		Status:             model.CardStatusActive,
		AuthorizedPersonID: &ap.AuthorizedPersonID,
		ExpiresAt:          time.Now().AddDate(3, 0, 0),
	}
	require.NoError(t, db.Create(card2).Error)
	apID := ap.AuthorizedPersonID

	countNil, err := repo.CountByAccountNumberAndAuthorizedPersonID(context.Background(), acc.AccountNumber, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countNil)

	countAP, err := repo.CountByAccountNumberAndAuthorizedPersonID(context.Background(), acc.AccountNumber, &apID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), countAP)
}

func TestRepoCard_CountNonDeactivatedByAccountNumberAndAuthorizedPersonID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewCardRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 83, curr.CurrencyID, 1000)

	activeCard := &model.Card{
		CardNumber:    fmt.Sprintf("4532000002%06d", uniqueCounter.Add(1)),
		CardType:      model.CardTypeDebit,
		CardBrand:     model.CardBrandVisa,
		Name:          "Active",
		AccountNumber: acc.AccountNumber,
		CVV:           "111",
		Limit:         500,
		Status:        model.CardStatusActive,
		ExpiresAt:     time.Now().AddDate(3, 0, 0),
	}
	deactivatedCard := &model.Card{
		CardNumber:    fmt.Sprintf("4532000003%06d", uniqueCounter.Add(1)),
		CardType:      model.CardTypeDebit,
		CardBrand:     model.CardBrandVisa,
		Name:          "Deactivated",
		AccountNumber: acc.AccountNumber,
		CVV:           "222",
		Limit:         500,
		Status:        model.CardStatusDeactivated,
		ExpiresAt:     time.Now().AddDate(3, 0, 0),
	}
	require.NoError(t, db.Create(activeCard).Error)
	require.NoError(t, db.Create(deactivatedCard).Error)

	count, err := repo.CountNonDeactivatedByAccountNumberAndAuthorizedPersonID(context.Background(), acc.AccountNumber, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRepoAuthorizedPerson_Create(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAuthorizedPersonRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 90, curr.CurrencyID, 2000)

	person := &model.AuthorizedPerson{
		AccountNumber: acc.AccountNumber,
		FirstName:     "Jane",
		LastName:      "Doe",
		DateOfBirth:   time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC),
		Gender:        "Female",
		Email:         "jane.doe@example.com",
		PhoneNumber:   "+381601234567",
		Address:       "123 Test Street",
	}

	err := repo.Create(context.Background(), person)
	require.NoError(t, err)
	assert.NotZero(t, person.AuthorizedPersonID)

	var loaded model.AuthorizedPerson
	require.NoError(t, db.First(&loaded, person.AuthorizedPersonID).Error)
	assert.Equal(t, "Jane", loaded.FirstName)
	assert.Equal(t, "Doe", loaded.LastName)
	assert.Equal(t, acc.AccountNumber, loaded.AccountNumber)
}

func TestRepoAuthorizedPerson_FindByID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAuthorizedPersonRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 91, curr.CurrencyID, 2000)

	person := &model.AuthorizedPerson{
		AccountNumber: acc.AccountNumber,
		FirstName:     "John",
		LastName:      "Smith",
		Email:         "john.smith@example.com",
	}
	require.NoError(t, db.Create(person).Error)

	found, err := repo.FindByID(context.Background(), person.AuthorizedPersonID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, person.AuthorizedPersonID, found.AuthorizedPersonID)
	assert.Equal(t, "John", found.FirstName)
}

func TestRepoAuthorizedPerson_FindByID_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAuthorizedPersonRepository(db)

	found, err := repo.FindByID(context.Background(), 99999)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRepoAuthorizedPerson_ListByAccountNumber(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAuthorizedPersonRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 92, curr.CurrencyID, 2000)
	other := seedAccount(t, db, 93, curr.CurrencyID, 500)

	p1 := &model.AuthorizedPerson{AccountNumber: acc.AccountNumber, FirstName: "Alice", LastName: "A", Email: "a@example.com"}
	p2 := &model.AuthorizedPerson{AccountNumber: acc.AccountNumber, FirstName: "Bob", LastName: "B", Email: "b@example.com"}
	pOther := &model.AuthorizedPerson{AccountNumber: other.AccountNumber, FirstName: "Charlie", LastName: "C", Email: "c@example.com"}
	require.NoError(t, db.Create(p1).Error)
	require.NoError(t, db.Create(p2).Error)
	require.NoError(t, db.Create(pOther).Error)

	people, err := repo.ListByAccountNumber(context.Background(), acc.AccountNumber)
	require.NoError(t, err)
	require.Len(t, people, 2)

	ids := []uint{people[0].AuthorizedPersonID, people[1].AuthorizedPersonID}
	assert.Contains(t, ids, p1.AuthorizedPersonID)
	assert.Contains(t, ids, p2.AuthorizedPersonID)
}

func TestRepoAuthorizedPerson_ListByAccountNumber_Empty(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewAuthorizedPersonRepository(db)

	curr := seedCurrency(t, db, model.RSD)
	acc := seedAccount(t, db, 94, curr.CurrencyID, 500)

	people, err := repo.ListByAccountNumber(context.Background(), acc.AccountNumber)
	require.NoError(t, err)
	assert.Empty(t, people)
}

func TestRepoExchangeRate_UpsertAll_Insert(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewExchangeRateRepository(db)

	now := time.Now()
	rates := []model.ExchangeRate{
		{
			CurrencyCode:         model.EUR,
			BaseCurrency:         model.RSD,
			BuyRate:              117.0,
			MiddleRate:           117.5,
			SellRate:             118.0,
			ProviderUpdatedAt:    now,
			ProviderNextUpdateAt: now.Add(2 * time.Hour),
		},
		{
			CurrencyCode:         model.USD,
			BaseCurrency:         model.RSD,
			BuyRate:              108.0,
			MiddleRate:           108.5,
			SellRate:             109.0,
			ProviderUpdatedAt:    now,
			ProviderNextUpdateAt: now.Add(2 * time.Hour),
		},
	}

	err := repo.UpsertAll(context.Background(), rates)
	require.NoError(t, err)

	all, err := repo.GetAll(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 2)

	codeSet := map[model.CurrencyCode]bool{}
	for _, r := range all {
		codeSet[r.CurrencyCode] = true
	}
	assert.True(t, codeSet[model.EUR])
	assert.True(t, codeSet[model.USD])
}

func TestRepoExchangeRate_UpsertAll_Update(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := repository.NewExchangeRateRepository(db)

	now := time.Now()
	initial := []model.ExchangeRate{
		{
			CurrencyCode:         model.GBP,
			BaseCurrency:         model.RSD,
			BuyRate:              135.0,
			MiddleRate:           136.0,
			SellRate:             137.0,
			ProviderUpdatedAt:    now,
			ProviderNextUpdateAt: now.Add(2 * time.Hour),
		},
	}
	require.NoError(t, repo.UpsertAll(context.Background(), initial))

	updated := []model.ExchangeRate{
		{
			CurrencyCode:         model.GBP,
			BaseCurrency:         model.RSD,
			BuyRate:              140.0,
			MiddleRate:           141.0,
			SellRate:             142.0,
			ProviderUpdatedAt:    now.Add(1 * time.Hour),
			ProviderNextUpdateAt: now.Add(3 * time.Hour),
		},
	}
	require.NoError(t, repo.UpsertAll(context.Background(), updated))

	all, err := repo.GetAll(context.Background())
	require.NoError(t, err)

	var gbpRate *model.ExchangeRate
	for i := range all {
		if all[i].CurrencyCode == model.GBP {
			gbpRate = &all[i]
			break
		}
	}
	require.NotNil(t, gbpRate)
	assert.Equal(t, 140.0, gbpRate.BuyRate)
	assert.Equal(t, 141.0, gbpRate.MiddleRate)
	assert.Equal(t, 142.0, gbpRate.SellRate)
}

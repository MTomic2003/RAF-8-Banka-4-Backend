//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitLoanRequest(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	seedBankAccounts(t, db, rsd.CurrencyID)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)
	loanType := seedLoanType(t, db)

	clientAuth := authHeaderForClient(t, 10, 100)

	testCases := []struct {
		name       string
		body       map[string]any
		auth       string
		wantStatus int
	}{
		{
			name: "happy path",
			body: map[string]any{
				"account_number":   account.AccountNumber,
				"loan_type_id":     loanType.LoanTypeID,
				"amount":           100000.0,
				"repayment_period": 24,
			},
			auth:       clientAuth,
			wantStatus: http.StatusCreated,
		},
		{
			name: "repayment period below min",
			body: map[string]any{
				"account_number":   account.AccountNumber,
				"loan_type_id":     loanType.LoanTypeID,
				"amount":           50000.0,
				"repayment_period": 1,
			},
			auth:       clientAuth,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repayment period above max",
			body: map[string]any{
				"account_number":   account.AccountNumber,
				"loan_type_id":     loanType.LoanTypeID,
				"amount":           50000.0,
				"repayment_period": 999,
			},
			auth:       clientAuth,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "loan type not found",
			body: map[string]any{
				"account_number":   account.AccountNumber,
				"loan_type_id":     99999,
				"amount":           50000.0,
				"repayment_period": 24,
			},
			auth:       clientAuth,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "account not found",
			body: map[string]any{
				"account_number":   "000000000000000000",
				"loan_type_id":     loanType.LoanTypeID,
				"amount":           50000.0,
				"repayment_period": 24,
			},
			auth:       clientAuth,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			body: map[string]any{
				"amount": 50000.0,
			},
			auth:       clientAuth,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing auth",
			body:       map[string]any{},
			auth:       "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", tc.body, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusCreated {
				resp := decodeResponse[map[string]any](t, recorder)
				assert.NotZero(t, resp["id"])
				assert.Equal(t, "PENDING", resp["status"])
				assert.NotZero(t, resp["monthly_installment"])
			}
		})
	}
}

func TestGetClientLoans(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	seedBankAccounts(t, db, rsd.CurrencyID)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)
	loanType := seedLoanType(t, db)

	clientAuth := authHeaderForClient(t, 10, 100)
	otherClientAuth := authHeaderForClient(t, 20, 200)

	performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
		"account_number":   account.AccountNumber,
		"loan_type_id":     loanType.LoanTypeID,
		"amount":           100000.0,
		"repayment_period": 24,
	}, clientAuth)

	testCases := []struct {
		name       string
		path       string
		auth       string
		wantStatus int
	}{
		{
			name:       "client gets own loans",
			path:       "/api/clients/100/loans",
			auth:       clientAuth,
			wantStatus: http.StatusOK,
		},
		{
			name:       "other client forbidden",
			path:       "/api/clients/100/loans",
			auth:       otherClientAuth,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing auth",
			path:       "/api/clients/100/loans",
			auth:       "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, tc.path, nil, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)
		})
	}
}

func TestApproveLoanRequest(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	seedBankAccounts(t, db, rsd.CurrencyID)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)
	loanType := seedLoanType(t, db)

	clientAuth := authHeaderForClient(t, 10, 100)
	employeeAuth := authHeaderForEmployee(t, 1, 1)

	createResp := performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
		"account_number":   account.AccountNumber,
		"loan_type_id":     loanType.LoanTypeID,
		"amount":           100000.0,
		"repayment_period": 24,
	}, clientAuth)
	requireStatus(t, createResp, http.StatusCreated)
	resp := decodeResponse[map[string]any](t, createResp)
	requestID := uint(resp["id"].(float64))

	approvePath := fmt.Sprintf("/api/loan-requests/%d/approve", requestID)

	recorder := performRequest(t, router, http.MethodPatch, approvePath, nil, clientAuth)
	requireStatus(t, recorder, http.StatusForbidden)

	recorder = performRequest(t, router, http.MethodPatch, approvePath, nil, employeeAuth)
	requireStatus(t, recorder, http.StatusOK)

	var loanRequest model.LoanRequest
	require.NoError(t, db.First(&loanRequest, requestID).Error)
	assert.Equal(t, model.LoanRequestApproved, loanRequest.Status)

	recorder = performRequest(t, router, http.MethodPatch, approvePath, nil, employeeAuth)
	requireStatus(t, recorder, http.StatusBadRequest)
}

func TestRejectLoanRequest(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)
	loanType := seedLoanType(t, db)

	clientAuth := authHeaderForClient(t, 10, 100)
	employeeAuth := authHeaderForEmployee(t, 1, 1)

	createResp := performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
		"account_number":   account.AccountNumber,
		"loan_type_id":     loanType.LoanTypeID,
		"amount":           50000.0,
		"repayment_period": 36,
	}, clientAuth)
	requireStatus(t, createResp, http.StatusCreated)
	resp := decodeResponse[map[string]any](t, createResp)
	requestID := uint(resp["id"].(float64))

	rejectPath := fmt.Sprintf("/api/loan-requests/%d/reject", requestID)

	recorder := performRequest(t, router, http.MethodPatch, rejectPath, nil, employeeAuth)
	requireStatus(t, recorder, http.StatusOK)

	var loanRequest model.LoanRequest
	require.NoError(t, db.First(&loanRequest, requestID).Error)
	assert.Equal(t, model.LoanRequestRejected, loanRequest.Status)

	recorder = performRequest(t, router, http.MethodPatch, rejectPath, nil, employeeAuth)
	requireStatus(t, recorder, http.StatusBadRequest)
}

func TestLoanRequestNotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	employeeAuth := authHeaderForEmployee(t, 1, 1)

	recorder := performRequest(t, router, http.MethodPatch, "/api/loan-requests/999999/approve", nil, employeeAuth)
	requireStatus(t, recorder, http.StatusNotFound)

	recorder = performRequest(t, router, http.MethodPatch, "/api/loan-requests/999999/reject", nil, employeeAuth)
	requireStatus(t, recorder, http.StatusNotFound)
}

func TestGetLoanByID(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	seedBankAccounts(t, db, rsd.CurrencyID)
	loanType := seedLoanType(t, db)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)

	clientAuth := authHeaderForClient(t, 10, 100)
	otherClientAuth := authHeaderForClient(t, 20, 200)
	employeeAuth := authHeaderForEmployee(t, 1, 1)

	createResp := performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
		"account_number":   account.AccountNumber,
		"loan_type_id":     loanType.LoanTypeID,
		"amount":           80000.0,
		"repayment_period": 24,
	}, clientAuth)
	requireStatus(t, createResp, http.StatusCreated)
	createRespBody := decodeResponse[map[string]any](t, createResp)
	requestID := uint(createRespBody["id"].(float64))

	approveResp := performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/loan-requests/%d/approve", requestID), nil, employeeAuth)
	requireStatus(t, approveResp, http.StatusOK)

	var loanReq model.LoanRequest
	require.NoError(t, db.Where("client_id = ?", 100).First(&loanReq).Error)
	var loan model.Loan
	require.NoError(t, db.Where("loan_request_id = ?", loanReq.ID).First(&loan).Error)
	loanID := loan.ID

	testCases := []struct {
		name       string
		path       string
		auth       string
		wantStatus int
	}{
		{
			name:       "happy path",
			path:       fmt.Sprintf("/api/clients/100/loans/%d", loanID),
			auth:       clientAuth,
			wantStatus: http.StatusOK,
		},
		{
			name:       "loan not found",
			path:       "/api/clients/100/loans/999999",
			auth:       clientAuth,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "other client forbidden",
			path:       fmt.Sprintf("/api/clients/100/loans/%d", loanID),
			auth:       otherClientAuth,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "missing auth",
			path:       fmt.Sprintf("/api/clients/100/loans/%d", loanID),
			auth:       "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, tc.path, nil, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusOK {
				resp := decodeResponse[map[string]any](t, recorder)
				assert.NotZero(t, resp["id"])
				assert.NotNil(t, resp["installments"])
			}
		})
	}
}

func TestGetClientLoansSorted(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	seedBankAccounts(t, db, rsd.CurrencyID)
	loanType := seedLoanType(t, db)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)

	clientAuth := authHeaderForClient(t, 10, 100)
	employeeAuth := authHeaderForEmployee(t, 1, 1)

	for _, amount := range []float64{20000.0, 50000.0, 35000.0} {
		createResp := performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
			"account_number":   account.AccountNumber,
			"loan_type_id":     loanType.LoanTypeID,
			"amount":           amount,
			"repayment_period": 24,
		}, clientAuth)
		requireStatus(t, createResp, http.StatusCreated)
		body := decodeResponse[map[string]any](t, createResp)
		reqID := uint(body["id"].(float64))
		performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/loan-requests/%d/approve", reqID), nil, employeeAuth)
	}

	recorderDesc := performRequest(t, router, http.MethodGet, "/api/clients/100/loans?sort=desc", nil, clientAuth)
	requireStatus(t, recorderDesc, http.StatusOK)

	recorderAsc := performRequest(t, router, http.MethodGet, "/api/clients/100/loans?sort=asc", nil, clientAuth)
	requireStatus(t, recorderAsc, http.StatusOK)

	recorderNone := performRequest(t, router, http.MethodGet, "/api/clients/100/loans", nil, clientAuth)
	requireStatus(t, recorderNone, http.StatusOK)
}

func TestListLoanRequestsFiltered(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)
	loanType := seedLoanType(t, db)

	clientAuth := authHeaderForClient(t, 10, 100)
	employeeAuth := authHeaderForEmployee(t, 1, 1)

	performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
		"account_number":   account.AccountNumber,
		"loan_type_id":     loanType.LoanTypeID,
		"amount":           40000.0,
		"repayment_period": 24,
	}, clientAuth)

	recorder := performRequest(t, router, http.MethodGet, "/api/loan-requests?page=1&page_size=5", nil, employeeAuth)
	requireStatus(t, recorder, http.StatusOK)
	resp := decodeResponse[map[string]any](t, recorder)
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(1), resp["page"])
	assert.Equal(t, float64(5), resp["page_size"])

	recorder = performRequest(t, router, http.MethodGet, "/api/loan-requests?status=PENDING", nil, employeeAuth)
	requireStatus(t, recorder, http.StatusOK)

	recorder = performRequest(t, router, http.MethodGet, "/api/loan-requests?client_id=100", nil, employeeAuth)
	requireStatus(t, recorder, http.StatusOK)
}

func TestListLoanRequests(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	rsd := seedCurrency(t, db, model.RSD)
	account := seedAccount(t, db, 100, rsd.CurrencyID, 50000)
	loanType := seedLoanType(t, db)

	clientAuth := authHeaderForClient(t, 10, 100)
	employeeAuth := authHeaderForEmployee(t, 1, 1)

	performRequest(t, router, http.MethodPost, "/api/clients/100/loans/request", map[string]any{
		"account_number":   account.AccountNumber,
		"loan_type_id":     loanType.LoanTypeID,
		"amount":           100000.0,
		"repayment_period": 24,
	}, clientAuth)

	recorder := performRequest(t, router, http.MethodGet, "/api/loan-requests?page=1&page_size=10", nil, employeeAuth)
	requireStatus(t, recorder, http.StatusOK)

	recorder = performRequest(t, router, http.MethodGet, "/api/loan-requests", nil, clientAuth)
	requireStatus(t, recorder, http.StatusForbidden)

	recorder = performRequest(t, router, http.MethodGet, "/api/loan-requests", nil, "")
	requireStatus(t, recorder, http.StatusUnauthorized)
}

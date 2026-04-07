//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExchangeRates(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	employeeAuth := authHeaderForEmployee(t, 1, 1)
	clientAuth := authHeaderForClient(t, 10, 100)

	seedExchangeRate(t, db, model.EUR, 116.0, 117.0, 118.0)
	seedExchangeRate(t, db, model.USD, 106.0, 107.0, 108.0)

	testCases := []struct {
		name       string
		auth       string
		wantStatus int
	}{
		{
			name:       "employee can get rates",
			auth:       employeeAuth,
			wantStatus: http.StatusOK,
		},
		{
			name:       "client can get rates",
			auth:       clientAuth,
			wantStatus: http.StatusOK,
		},
		{
			name:       "no auth still works (public endpoint)",
			auth:       "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, "/api/exchange/rates", nil, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusOK {
				resp := decodeResponse[map[string]any](t, recorder)
				rates, ok := resp["rates"].([]any)
				require.True(t, ok)
				assert.GreaterOrEqual(t, len(rates), 2)
				assert.Equal(t, "RSD", resp["base_currency"])
			}
		})
	}
}

func TestGetExchangeRatesNoData(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	recorder := performRequest(t, router, http.MethodGet, "/api/exchange/rates", nil, "")
	requireStatus(t, recorder, http.StatusServiceUnavailable)
}

func TestConvertCurrency(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	seedExchangeRate(t, db, model.EUR, 116.0, 117.0, 118.0)

	clientAuth := authHeaderForClient(t, 10, 100)

	testCases := []struct {
		name       string
		path       string
		auth       string
		wantStatus int
	}{
		{
			name:       "happy path",
			path:       "/api/exchange/calculate?amount=100&from_currency=EUR&to_currency=RSD",
			auth:       clientAuth,
			wantStatus: http.StatusOK,
		},
		{
			name:       "same currency",
			path:       "/api/exchange/calculate?amount=100&from_currency=RSD&to_currency=RSD",
			auth:       clientAuth,
			wantStatus: http.StatusOK,
		},
		{
			name:       "no auth still works (public endpoint)",
			path:       "/api/exchange/calculate?amount=100&from_currency=EUR&to_currency=RSD",
			auth:       "",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, tc.path, nil, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusOK {
				resp := decodeResponse[map[string]any](t, recorder)
				assert.NotZero(t, resp["total"])
			}
		})
	}
}

func TestConvertCurrencyValidation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	seedExchangeRate(t, db, model.EUR, 116.0, 117.0, 118.0)

	testCases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "missing amount",
			path:       "/api/exchange/calculate?from_currency=EUR&to_currency=RSD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing from_currency",
			path:       "/api/exchange/calculate?amount=100&to_currency=RSD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing to_currency",
			path:       "/api/exchange/calculate?amount=100&from_currency=EUR",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero amount",
			path:       "/api/exchange/calculate?amount=0&from_currency=EUR&to_currency=RSD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "negative amount",
			path:       "/api/exchange/calculate?amount=-50&from_currency=EUR&to_currency=RSD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "non-numeric amount",
			path:       "/api/exchange/calculate?amount=abc&from_currency=EUR&to_currency=RSD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported from_currency",
			path:       "/api/exchange/calculate?amount=100&from_currency=XYZ&to_currency=RSD",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unsupported to_currency",
			path:       "/api/exchange/calculate?amount=100&from_currency=EUR&to_currency=XYZ",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, tc.path, nil, "")
			requireStatus(t, recorder, tc.wantStatus)
		})
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	recorder := performRequest(t, router, http.MethodGet, "/api/health", nil, "")
	requireStatus(t, recorder, http.StatusOK)

	resp := decodeResponse[map[string]any](t, recorder)
	assert.Equal(t, "OK", resp["status"])
}

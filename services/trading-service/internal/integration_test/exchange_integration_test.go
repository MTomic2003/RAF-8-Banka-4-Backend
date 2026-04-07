//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodGet, "/api/health", nil, "")
	requireStatus(t, rec, http.StatusOK)
}

func TestGetAllExchanges(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	seedExchange(t, db, "XNYS")
	seedExchange(t, db, "XNAS")

	rec := performRequest(t, router, http.MethodGet, "/api/exchanges?page=1&page_size=10", nil, "")
	requireStatus(t, rec, http.StatusOK)

	body := decodeResponse[map[string]any](t, rec)
	data := body["data"].([]any)
	require.GreaterOrEqual(t, len(data), 2)
	require.Equal(t, float64(2), body["total"])
}

func TestGetAllExchanges_DefaultPagination(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	seedExchange(t, db, "XLON")

	rec := performRequest(t, router, http.MethodGet, "/api/exchanges", nil, "")
	requireStatus(t, rec, http.StatusOK)
}

func TestToggleTradingEnabled(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XTST")
	require.True(t, ex.TradingEnabled)

	rec := performRequest(t, router, http.MethodPatch, "/api/exchanges/XTST/toggle", nil, "")
	requireStatus(t, rec, http.StatusOK)

	body := decodeResponse[map[string]any](t, rec)
	require.Equal(t, false, body["trading_enabled"])

	rec = performRequest(t, router, http.MethodPatch, "/api/exchanges/XTST/toggle", nil, "")
	requireStatus(t, rec, http.StatusOK)

	body = decodeResponse[map[string]any](t, rec)
	require.Equal(t, true, body["trading_enabled"])
}

func TestToggleTradingEnabled_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodPatch, "/api/exchanges/NOPE/toggle", nil, "")
	require.NotEqual(t, http.StatusOK, rec.Code)
}

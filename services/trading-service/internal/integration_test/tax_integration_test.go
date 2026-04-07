//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListTaxUsers(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/tax?page=1&page_size=10", nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	data := resp["data"].([]any)
	require.NotNil(t, data)
}

func TestListTaxUsers_FilterByClient(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/tax?userType=client", nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	data := resp["data"].([]any)
	for _, entry := range data {
		e := entry.(map[string]any)
		require.Equal(t, "client", e["userType"])
	}
}

func TestListTaxUsers_FilterByActuary(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/tax?userType=actuary", nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	data := resp["data"].([]any)
	for _, entry := range data {
		e := entry.(map[string]any)
		require.Equal(t, "actuary", e["userType"])
	}
}

func TestListTaxUsers_DefaultPagination(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/tax", nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestListTaxUsers_ForbiddenForAgent(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForAgent(t)

	rec := performRequest(t, router, http.MethodGet, "/api/tax", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestCollectTaxes(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPost, "/api/tax/collect", nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	require.Equal(t, "Tax collection completed", resp["message"])
}

func TestCollectTaxes_ForbiddenForClient(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForClient(t, 50, 1)

	rec := performRequest(t, router, http.MethodPost, "/api/tax/collect", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

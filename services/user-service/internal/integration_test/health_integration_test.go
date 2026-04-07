//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	recorder := performRequest(t, router, http.MethodGet, "/api/health", nil, "")
	requireStatus(t, recorder, http.StatusOK)

	type healthResponse struct {
		Status string `json:"status"`
	}

	response := decodeResponse[healthResponse](t, recorder)
	assert.Equal(t, "OK", response.Status)
}

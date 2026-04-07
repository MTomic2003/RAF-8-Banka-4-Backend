//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthLoginClaims(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, employee := seedEmployee(t, db, position.PositionID)

	recorder := performRequest(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    identity.Email,
		"password": "Password12",
	}, "")

	requireStatus(t, recorder, http.StatusOK)

	response := decodeResponse[loginResponse](t, recorder)
	assert.Equal(t, employee.EmployeeID, response.User.ID)

	claims := verifyAccessToken(t, response.Token)
	assert.Equal(t, identity.ID, claims.IdentityID)
	assert.Equal(t, string(auth.IdentityEmployee), claims.IdentityType)
	if assert.NotNil(t, claims.EmployeeID) {
		assert.Equal(t, employee.EmployeeID, *claims.EmployeeID)
	}
	assert.Nil(t, claims.ClientID)
}

func TestAuthRefreshClaims(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, employee := seedEmployee(t, db, position.PositionID)

	login := loginEmployee(t, router, identity.Email, "Password12")

	recorder := performRequest(t, router, http.MethodPost, "/api/auth/refresh", map[string]any{
		"refresh_token": login.RefreshToken,
	}, "")

	requireStatus(t, recorder, http.StatusOK)

	response := decodeResponse[loginResponse](t, recorder)
	assert.Equal(t, employee.EmployeeID, response.User.ID)

	claims := verifyAccessToken(t, response.Token)
	assert.Equal(t, identity.ID, claims.IdentityID)
	assert.Equal(t, string(auth.IdentityEmployee), claims.IdentityType)
	if assert.NotNil(t, claims.EmployeeID) {
		assert.Equal(t, employee.EmployeeID, *claims.EmployeeID)
	}
	assert.Nil(t, claims.ClientID)
}

func TestResendActivation(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, _ := seedEmployee(t, db, position.PositionID)
	identity.Active = false
	identity.PasswordHash = ""
	require.NoError(t, db.Save(identity).Error)

	testCases := []struct {
		name       string
		body       any
		rawBody    string
		wantStatus int
	}{
		{
			name:       "existing inactive account",
			body:       map[string]any{"email": identity.Email},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown email returns 200 to prevent enumeration",
			body:       map[string]any{"email": "nobody@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			rawBody:    "{bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.rawBody != "" {
				r := performRawJSONRequest(t, router, http.MethodPost, "/api/auth/resend-activation", tc.rawBody, "")
				requireStatus(t, r, tc.wantStatus)
			} else {
				r := performRequest(t, router, http.MethodPost, "/api/auth/resend-activation", tc.body, "")
				requireStatus(t, r, tc.wantStatus)
			}
		})
	}
}

func TestForgotPasswordErrors(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	recorder := performRawJSONRequest(t, router, http.MethodPost, "/api/auth/forgot-password", "{bad", "")
	requireStatus(t, recorder, http.StatusBadRequest)
}

func TestResetPasswordErrors(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	invalidJSON := performRawJSONRequest(t, router, http.MethodPost, "/api/auth/reset-password", "{bad", "")
	requireStatus(t, invalidJSON, http.StatusBadRequest)

	expiredToken := performRequest(t, router, http.MethodPost, "/api/auth/reset-password", map[string]any{
		"token":        "nonexistent-token",
		"new_password": "Password12",
	}, "")
	requireStatus(t, expiredToken, http.StatusBadRequest)
}

func TestChangePassword(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, _ := seedEmployee(t, db, position.PositionID)

	login := loginEmployee(t, router, identity.Email, "Password12")
	bearer := "Bearer " + login.Token

	testCases := []struct {
		name       string
		body       any
		rawBody    string
		auth       string
		wantStatus int
	}{
		{
			name: "valid password change",
			body: map[string]any{
				"old_password": "Password12",
				"new_password": "NewPassword34",
			},
			auth:       bearer,
			wantStatus: http.StatusOK,
		},
		{
			name: "wrong old password",
			body: map[string]any{
				"old_password": "WrongPassword",
				"new_password": "AnotherPassword56",
			},
			auth:       bearer,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing auth header",
			body:       map[string]any{"old_password": "Password12", "new_password": "NewPassword34"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid json",
			rawBody:    "{bad",
			auth:       bearer,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.rawBody != "" {
				recorder := performRawJSONRequest(t, router, http.MethodPost, "/api/auth/change-password", tc.rawBody, tc.auth)
				requireStatus(t, recorder, tc.wantStatus)
			} else {
				recorder := performRequest(t, router, http.MethodPost, "/api/auth/change-password", tc.body, tc.auth)
				requireStatus(t, recorder, tc.wantStatus)
			}
		})
	}
}

func TestRefreshTokenErrors(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)

	invalidJSON := performRawJSONRequest(t, router, http.MethodPost, "/api/auth/refresh", "{bad", "")
	requireStatus(t, invalidJSON, http.StatusBadRequest)

	missingToken := performRequest(t, router, http.MethodPost, "/api/auth/refresh", map[string]any{
		"refresh_token": "bogus-token-value",
	}, "")
	requireStatus(t, missingToken, http.StatusUnauthorized)
}

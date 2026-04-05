//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	commonpermission "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, employee := seedEmployee(t, db, position.PositionID)

	testCases := []struct {
		name       string
		body       any
		rawBody    string
		wantStatus int
		assertBody func(t *testing.T, recorderBody loginResponse)
	}{
		{
			name: "correct credentials",
			body: map[string]any{
				"email":    identity.Email,
				"password": "Password12",
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, response loginResponse) {
				t.Helper()
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, employee.EmployeeID, response.User.ID)
			},
		},
		{
			name: "wrong password",
			body: map[string]any{
				"email":    identity.Email,
				"password": "WrongPassword12",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "non existent email",
			body: map[string]any{
				"email":    "missing@example.com",
				"password": "Password12",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid json",
			rawBody:    "{",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var recorder *httptest.ResponseRecorder
			if tc.rawBody != "" {
				recorder = performRawJSONRequest(t, router, http.MethodPost, "/api/auth/login", tc.rawBody, "")
			} else {
				recorder = performRequest(t, router, http.MethodPost, "/api/auth/login", tc.body, "")
			}

			requireStatus(t, recorder, tc.wantStatus)
			if tc.assertBody != nil {
				tc.assertBody(t, decodeResponse[loginResponse](t, recorder))
			}
		})
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	adminIdentity, _ := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeCreate)
	noPermIdentity, _ := seedEmployee(t, db, position.PositionID)

	validBody := map[string]any{
		"first_name":    "Ana",
		"last_name":     "Test",
		"gender":        "female",
		"date_of_birth": time.Now().UTC().AddDate(-25, 0, 0).Format(time.RFC3339),
		"email":         uniqueValue(t, "register") + "@example.com",
		"phone_number":  "0600000000",
		"address":       "Main 1",
		"username":      uniqueValue(t, "register-user"),
		"department":    "QA",
		"position_id":   position.PositionID,
		"active":        false,
		"permissions":   []string{string(commonpermission.EmployeeView)},
	}

	testCases := []struct {
		name       string
		body       map[string]any
		auth       string
		wantStatus int
		wantEmail  string
	}{
		{
			name:       "valid data with permission",
			body:       validBody,
			auth:       authHeader(t, adminIdentity.ID),
			wantStatus: http.StatusCreated,
			wantEmail:  validBody["email"].(string),
		},
		{
			name: "duplicate email",
			body: map[string]any{
				"first_name":    "Ana",
				"last_name":     "Dup",
				"gender":        "female",
				"date_of_birth": time.Now().UTC().AddDate(-24, 0, 0).Format(time.RFC3339),
				"email":         validBody["email"],
				"phone_number":  "0600000001",
				"address":       "Main 2",
				"username":      uniqueValue(t, "register-user"),
				"department":    "QA",
				"position_id":   position.PositionID,
				"active":        true,
			},
			auth:       authHeader(t, adminIdentity.ID),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "missing auth header",
			body:       map[string]any{"email": uniqueValue(t, "missing-auth") + "@example.com"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "without employee create permission",
			body: map[string]any{
				"first_name":    "Ana",
				"last_name":     "NoPerm",
				"gender":        "female",
				"date_of_birth": time.Now().UTC().AddDate(-24, 0, 0).Format(time.RFC3339),
				"email":         uniqueValue(t, "noperm") + "@example.com",
				"phone_number":  "0600000002",
				"address":       "Main 3",
				"username":      uniqueValue(t, "noperm-user"),
				"department":    "QA",
				"position_id":   position.PositionID,
				"active":        true,
			},
			auth:       authHeader(t, noPermIdentity.ID),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodPost, "/api/employees/register", tc.body, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusCreated {
				response := decodeResponse[employeeResponse](t, recorder)
				assert.Equal(t, tc.wantEmail, response.Email)
				assert.NotZero(t, response.ID)
			}
		})
	}
}

func TestListEmployees(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	developer := seedPosition(t, db)
	developer.Title = "Developer"
	require.NoError(t, db.Save(developer).Error)

	manager := seedPosition(t, db)
	manager.Title = "Manager"
	require.NoError(t, db.Save(manager).Error)

	viewerIdentity, _ := seedEmployeeWithPermissions(t, db, developer.PositionID, commonpermission.EmployeeView)
	devOneIdentity, _ := seedEmployee(t, db, developer.PositionID)
	devTwoIdentity, _ := seedEmployee(t, db, developer.PositionID)
	managerIdentity, _ := seedEmployee(t, db, manager.PositionID)

	testCases := []struct {
		name         string
		path         string
		auth         string
		wantStatus   int
		wantTotal    int64
		wantCount    int
		assertResult func(t *testing.T, response listEmployeesResponse)
	}{
		{
			name:       "list all employees",
			path:       "/api/employees?page=1&page_size=10",
			auth:       authHeader(t, viewerIdentity.ID),
			wantStatus: http.StatusOK,
			wantTotal:  4,
			wantCount:  4,
		},
		{
			name:       "filter by email",
			path:       "/api/employees?email=" + url.QueryEscape(devOneIdentity.Email) + "&page=1&page_size=10",
			auth:       authHeader(t, viewerIdentity.ID),
			wantStatus: http.StatusOK,
			wantTotal:  1,
			wantCount:  1,
			assertResult: func(t *testing.T, response listEmployeesResponse) {
				t.Helper()
				assert.Equal(t, devOneIdentity.Email, response.Data[0].Email)
			},
		},
		{
			name:       "filter by position title",
			path:       "/api/employees?position=Developer&page=1&page_size=10",
			auth:       authHeader(t, viewerIdentity.ID),
			wantStatus: http.StatusOK,
			wantTotal:  3,
			wantCount:  3,
			assertResult: func(t *testing.T, response listEmployeesResponse) {
				t.Helper()
				emails := []string{response.Data[0].Email, response.Data[1].Email, response.Data[2].Email}
				assert.Contains(t, emails, viewerIdentity.Email)
				assert.Contains(t, emails, devOneIdentity.Email)
				assert.Contains(t, emails, devTwoIdentity.Email)
				assert.NotContains(t, emails, managerIdentity.Email)
			},
		},
		{
			name:       "missing auth header",
			path:       "/api/employees",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, tc.path, nil, tc.auth)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusOK {
				response := decodeResponse[listEmployeesResponse](t, recorder)
				assert.Equal(t, tc.wantTotal, response.Total)
				assert.Len(t, response.Data, tc.wantCount)
				if tc.assertResult != nil {
					tc.assertResult(t, response)
				}
			}
		})
	}
}

func TestGetEmployee(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	viewerIdentity, _ := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeView)
	empIdentity, employee := seedEmployee(t, db, position.PositionID)
	authHdr := authHeader(t, viewerIdentity.ID)

	testCases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "valid id", path: "/api/employees/" + itoa(employee.EmployeeID), wantStatus: http.StatusOK},
		{name: "missing employee", path: "/api/employees/999", wantStatus: http.StatusNotFound},
		{name: "invalid id format", path: "/api/employees/abc", wantStatus: http.StatusBadRequest},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodGet, tc.path, nil, authHdr)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusOK {
				response := decodeResponse[employeeResponse](t, recorder)
				assert.Equal(t, empIdentity.Email, response.Email)
				assert.Equal(t, employee.EmployeeID, response.ID)
			}
		})
	}
}

func TestUpdateEmployee(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	otherPosition := seedPosition(t, db)
	updaterIdentity, _ := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeUpdate)
	_, employee := seedEmployee(t, db, position.PositionID)
	otherIdentity, _ := seedEmployee(t, db, position.PositionID)
	authHdr := authHeader(t, updaterIdentity.ID)

	newDepartment := "Operations"
	newPositionID := otherPosition.PositionID

	testCases := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name: "valid partial update",
			body: map[string]any{
				"department":  newDepartment,
				"position_id": newPositionID,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "duplicate email",
			body: map[string]any{
				"email": otherIdentity.Email,
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid position id",
			body: map[string]any{
				"position_id": uint(9999),
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodPatch, "/api/employees/"+itoa(employee.EmployeeID), tc.body, authHdr)
			requireStatus(t, recorder, tc.wantStatus)

			if tc.wantStatus == http.StatusOK {
				response := decodeResponse[employeeResponse](t, recorder)
				assert.Equal(t, newDepartment, response.Department)
				assert.Equal(t, newPositionID, response.PositionID)
			}
		})
	}
}

func TestActivateAccount(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, _ := seedEmployee(t, db, position.PositionID)
	identity.Active = false
	identity.PasswordHash = ""
	require.NoError(t, db.Save(identity).Error)

	validToken := &model.ActivationToken{
		IdentityID: identity.ID,
		Token:      uniqueValue(t, "activation-token"),
		ExpiresAt:  time.Now().Add(30 * time.Minute),
	}
	expiredToken := &model.ActivationToken{
		IdentityID: identity.ID,
		Token:      uniqueValue(t, "activation-token"),
		ExpiresAt:  time.Now().Add(-30 * time.Minute),
	}
	require.NoError(t, db.Create(validToken).Error)
	require.NoError(t, db.Create(expiredToken).Error)

	testCases := []struct {
		name       string
		body       map[string]any
		wantStatus int
	}{
		{
			name: "valid token and password",
			body: map[string]any{
				"token":    validToken.Token,
				"password": "Password12",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "expired token",
			body: map[string]any{
				"token":    expiredToken.Token,
				"password": "Password12",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "weak password",
			body: map[string]any{
				"token":    validToken.Token,
				"password": "weak",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodPost, "/api/auth/activate", tc.body, "")
			requireStatus(t, recorder, tc.wantStatus)
		})
	}

	refreshedIdentity := &model.Identity{}
	require.NoError(t, db.First(refreshedIdentity, identity.ID).Error)
	assert.NotEmpty(t, refreshedIdentity.PasswordHash)
}

func TestPasswordResetFlow(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, _ := seedEmployee(t, db, position.PositionID)

	forgotPassword := performRequest(t, router, http.MethodPost, "/api/auth/forgot-password", map[string]any{
		"email": identity.Email,
	}, "")
	requireStatus(t, forgotPassword, http.StatusOK)

	var resetToken model.ResetToken
	require.NoError(t, db.WithContext(context.Background()).Where("identity_id = ?", identity.ID).First(&resetToken).Error)

	resetPassword := performRequest(t, router, http.MethodPost, "/api/auth/reset-password", map[string]any{
		"token":        resetToken.Token,
		"new_password": "NewPassword12",
	}, "")
	requireStatus(t, resetPassword, http.StatusOK)

	login := performRequest(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    identity.Email,
		"password": "NewPassword12",
	}, "")
	requireStatus(t, login, http.StatusOK)
}

func TestRefreshToken(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	identity, _ := seedEmployee(t, db, position.PositionID)

	login := loginEmployee(t, router, identity.Email, "Password12")

	validRefresh := performRequest(t, router, http.MethodPost, "/api/auth/refresh", map[string]any{
		"refresh_token": login.RefreshToken,
	}, "")
	requireStatus(t, validRefresh, http.StatusOK)

	refreshResponse := decodeResponse[loginResponse](t, validRefresh)
	assert.NotEmpty(t, refreshResponse.Token)
	assert.NotEmpty(t, refreshResponse.RefreshToken)
	assert.NotEqual(t, login.RefreshToken, refreshResponse.RefreshToken)

	invalidRefresh := performRequest(t, router, http.MethodPost, "/api/auth/refresh", map[string]any{
		"refresh_token": "invalid-token",
	}, "")
	requireStatus(t, invalidRefresh, http.StatusUnauthorized)
}

func itoa(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

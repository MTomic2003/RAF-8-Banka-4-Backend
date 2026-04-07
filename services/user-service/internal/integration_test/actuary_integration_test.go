//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	commonpermission "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterEmployeeAsAgent(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)
	adminIdentity, admin := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.All...)

	recorder := performRequest(t, router, http.MethodPost, "/api/employees/register", map[string]any{
		"first_name":    "Agent",
		"last_name":     "User",
		"gender":        "female",
		"date_of_birth": time.Now().UTC().AddDate(-28, 0, 0).Format(time.RFC3339),
		"email":         uniqueValue(t, "agent") + "@example.com",
		"phone_number":  "0600000009",
		"address":       "Main 9",
		"username":      uniqueValue(t, "agent-user"),
		"department":    "Trading",
		"position_id":   position.PositionID,
		"active":        true,
		"is_agent":      true,
		"limit":         100000.0,
		"need_approval": true,
	}, authHeader(t, adminIdentity.ID, admin.EmployeeID))

	requireStatus(t, recorder, http.StatusCreated)

	response := decodeResponse[employeeResponse](t, recorder)
	assert.True(t, response.IsAgent)
	assert.False(t, response.IsSupervisor)
	assert.Equal(t, 100000.0, response.Limit)
	assert.True(t, response.NeedApproval)
}

func TestListActuariesAndManageAgent(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)

	viewerIdentity, viewer := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeView)
	supervisorIdentity, supervisor := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeUpdate)
	_, agent := seedEmployee(t, db, position.PositionID)

	require.NoError(t, db.Create(&model.ActuaryInfo{
		EmployeeID:   supervisor.EmployeeID,
		IsSupervisor: true,
	}).Error)
	require.NoError(t, db.Create(&model.ActuaryInfo{
		EmployeeID:   agent.EmployeeID,
		IsAgent:      true,
		Limit:        75000,
		UsedLimit:    5000,
		NeedApproval: true,
	}).Error)

	listRecorder := performRequest(
		t,
		router,
		http.MethodGet,
		"/api/actuaries?type=agent&page=1&page_size=10",
		nil,
		authHeader(t, viewerIdentity.ID, viewer.EmployeeID),
	)

	requireStatus(t, listRecorder, http.StatusOK)
	listResponse := decodeResponse[listActuariesResponse](t, listRecorder)
	require.Len(t, listResponse.Data, 1)
	assert.Equal(t, agent.EmployeeID, listResponse.Data[0].ID)
	assert.True(t, listResponse.Data[0].IsAgent)

	forbiddenUpdateRecorder := performRequest(
		t,
		router,
		http.MethodPatch,
		"/api/actuaries/"+itoa(agent.EmployeeID),
		map[string]any{
			"limit": 90000.0,
		},
		authHeader(t, viewerIdentity.ID, viewer.EmployeeID),
	)

	requireStatus(t, forbiddenUpdateRecorder, http.StatusForbidden)

	updateRecorder := performRequest(
		t,
		router,
		http.MethodPatch,
		"/api/actuaries/"+itoa(agent.EmployeeID),
		map[string]any{
			"limit":         90000.0,
			"need_approval": false,
		},
		authHeader(t, supervisorIdentity.ID, supervisor.EmployeeID),
	)

	requireStatus(t, updateRecorder, http.StatusOK)
	updateResponse := decodeResponse[actuaryResponse](t, updateRecorder)
	assert.Equal(t, 90000.0, updateResponse.Limit)
	assert.False(t, updateResponse.NeedApproval)

	resetRecorder := performRequest(
		t,
		router,
		http.MethodPost,
		"/api/actuaries/"+itoa(agent.EmployeeID)+"/reset-used-limit",
		nil,
		authHeader(t, supervisorIdentity.ID, supervisor.EmployeeID),
	)

	requireStatus(t, resetRecorder, http.StatusOK)
	resetResponse := decodeResponse[actuaryResponse](t, resetRecorder)
	assert.Zero(t, resetResponse.UsedLimit)
}

func TestListActuariesPagination(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)

	viewerIdentity, viewer := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeView)

	for i := 0; i < 5; i++ {
		_, emp := seedEmployee(t, db, position.PositionID)
		require.NoError(t, db.Create(&model.ActuaryInfo{
			EmployeeID: emp.EmployeeID,
			IsAgent:    true,
			Limit:      float64((i + 1) * 10000),
		}).Error)
	}

	validAuth := authHeader(t, viewerIdentity.ID, viewer.EmployeeID)

	testCases := []struct {
		name         string
		path         string
		auth         string
		wantStatus   int
		wantCount    int
		wantTotalGte int64
	}{
		{
			name:         "page 1 of agents",
			path:         "/api/actuaries?type=agent&page=1&page_size=3",
			auth:         validAuth,
			wantStatus:   http.StatusOK,
			wantCount:    3,
			wantTotalGte: 5,
		},
		{
			name:         "page 2 of agents",
			path:         "/api/actuaries?type=agent&page=2&page_size=3",
			auth:         validAuth,
			wantStatus:   http.StatusOK,
			wantCount:    2,
			wantTotalGte: 5,
		},
		{
			name:       "filter by type supervisor returns none",
			path:       "/api/actuaries?type=supervisor&page=1&page_size=10",
			auth:       validAuth,
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name:       "missing auth",
			path:       "/api/actuaries?page=1&page_size=10",
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
				response := decodeResponse[listActuariesResponse](t, recorder)
				assert.Len(t, response.Data, tc.wantCount)
				if tc.wantTotalGte > 0 {
					assert.GreaterOrEqual(t, response.Total, tc.wantTotalGte)
				}
			}
		})
	}
}

func TestUpdateActuarySettingsErrors(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)

	supervisorIdentity, supervisor := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeUpdate)
	require.NoError(t, db.Create(&model.ActuaryInfo{
		EmployeeID:   supervisor.EmployeeID,
		IsSupervisor: true,
	}).Error)

	testCases := []struct {
		name       string
		path       string
		body       any
		rawBody    string
		wantStatus int
	}{
		{
			name:       "invalid id format",
			path:       "/api/actuaries/abc",
			body:       map[string]any{"limit": 1000.0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "agent not found",
			path:       fmt.Sprintf("/api/actuaries/%d", 999999),
			body:       map[string]any{"limit": 1000.0},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid json body",
			path:       "/api/actuaries/" + itoa(supervisor.EmployeeID),
			rawBody:    "{bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.rawBody != "" {
				r := performRawJSONRequest(t, router, http.MethodPatch, tc.path, tc.rawBody, authHeader(t, supervisorIdentity.ID, supervisor.EmployeeID))
				requireStatus(t, r, tc.wantStatus)
			} else {
				r := performRequest(t, router, http.MethodPatch, tc.path, tc.body, authHeader(t, supervisorIdentity.ID, supervisor.EmployeeID))
				requireStatus(t, r, tc.wantStatus)
			}
		})
	}
}

func TestResetUsedLimitErrors(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	router := setupTestRouter(t, db)
	position := seedPosition(t, db)

	supervisorIdentity, supervisor := seedEmployeeWithPermissions(t, db, position.PositionID, commonpermission.EmployeeUpdate)
	require.NoError(t, db.Create(&model.ActuaryInfo{
		EmployeeID:   supervisor.EmployeeID,
		IsSupervisor: true,
	}).Error)

	testCases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "invalid id format",
			path:       "/api/actuaries/abc/reset-used-limit",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "agent not found",
			path:       "/api/actuaries/999999/reset-used-limit",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			recorder := performRequest(t, router, http.MethodPost, tc.path, nil, authHeader(t, supervisorIdentity.ID, supervisor.EmployeeID))
			requireStatus(t, recorder, tc.wantStatus)
		})
	}
}

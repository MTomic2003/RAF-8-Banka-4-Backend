//go:build integration

package integration_test

import (
	"net/http"
	"testing"
	"time"

	commonpermission "common/pkg/permission"
	"user-service/internal/model"

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

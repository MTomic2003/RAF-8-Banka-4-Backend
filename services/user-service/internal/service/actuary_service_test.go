package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

func TestUpdateActuarySettings(t *testing.T) {
	t.Parallel()

	agent := activeAgent()

	repo := &fakeActuaryRepo{
		byEmployeeID: map[uint]*model.ActuaryInfo{
			agent.EmployeeID: agent.ActuaryInfo,
		},
	}

	service := NewActuaryService(
		repo,
		&fakeEmployeeRepo{
			byIDs: map[uint]*model.Employee{
				agent.EmployeeID: agent,
			},
		},
	)

	response, err := service.UpdateActuarySettings(
		context.Background(),
		agent.EmployeeID,
		&dto.UpdateActuarySettingsRequest{
			Limit:        ptr(200000.0),
			NeedApproval: ptr(false),
		},
	)

	require.NoError(t, err)
	require.Equal(t, 200000.0, response.Limit)
	require.False(t, response.NeedApproval)
}

func TestResetUsedLimit(t *testing.T) {
	t.Parallel()

	agent := activeAgent()

	repo := &fakeActuaryRepo{
		byEmployeeID: map[uint]*model.ActuaryInfo{
			agent.EmployeeID: agent.ActuaryInfo,
		},
	}

	service := NewActuaryService(
		repo,
		&fakeEmployeeRepo{
			byIDs: map[uint]*model.Employee{
				agent.EmployeeID: agent,
			},
		},
	)

	response, err := service.ResetUsedLimit(
		context.Background(),
		agent.EmployeeID,
	)

	require.NoError(t, err)
	require.Zero(t, response.UsedLimit)
}

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

// ── GetAll Tests ─────────────────────────────────────────────────

func TestExchangeService_GetAll_Success(t *testing.T) {
	exchanges := []model.Exchange{
		{ExchangeID: 1, Name: "NYSE", MicCode: "XNYS"},
		{ExchangeID: 2, Name: "Nasdaq", MicCode: "XNAS"},
	}
	repo := &fakeExchangeRepo{exchanges: exchanges, total: 2}
	svc := NewExchangeService(repo)

	result, total, err := svc.GetAll(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int64(2), total)
	require.Equal(t, "XNYS", result[0].MicCode)
	require.Equal(t, "XNAS", result[1].MicCode)
}

func TestExchangeService_GetAll_Empty(t *testing.T) {
	repo := &fakeExchangeRepo{exchanges: []model.Exchange{}, total: 0}
	svc := NewExchangeService(repo)

	result, total, err := svc.GetAll(context.Background(), 0, 10)
	require.NoError(t, err)
	require.Empty(t, result)
	require.Equal(t, int64(0), total)
}

func TestExchangeService_GetAll_RepoError(t *testing.T) {
	repo := &fakeExchangeRepo{findAllErr: errors.New("db connection failed")}
	svc := NewExchangeService(repo)

	result, total, err := svc.GetAll(context.Background(), 0, 10)
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, int64(0), total)
}

// ── ToggleTradingEnabled Tests ───────────────────────────────────

func TestExchangeService_ToggleTradingEnabled_Success(t *testing.T) {
	toggled := &model.Exchange{ExchangeID: 1, Name: "NYSE", MicCode: "XNYS", TradingEnabled: false}
	repo := &fakeExchangeRepo{toggledExch: toggled}
	svc := NewExchangeService(repo)

	result, err := svc.ToggleTradingEnabled(context.Background(), "XNYS")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "XNYS", result.MicCode)
	require.False(t, result.TradingEnabled)
}

func TestExchangeService_ToggleTradingEnabled_NotFound(t *testing.T) {
	repo := &fakeExchangeRepo{toggledExch: nil, toggleErr: nil}
	svc := NewExchangeService(repo)

	result, err := svc.ToggleTradingEnabled(context.Background(), "XXXX")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "exchange not found")
}

func TestExchangeService_ToggleTradingEnabled_RepoError(t *testing.T) {
	repo := &fakeExchangeRepo{toggleErr: errors.New("db error")}
	svc := NewExchangeService(repo)

	result, err := svc.ToggleTradingEnabled(context.Background(), "XNYS")
	require.Error(t, err)
	require.Nil(t, result)
}

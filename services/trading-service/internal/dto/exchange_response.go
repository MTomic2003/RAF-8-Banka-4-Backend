package dto

import "github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"

type ExchangeResponse struct {
	ExchangeID     uint   `json:"exchange_id"`
	Name           string `json:"name"`
	Acronym        string `json:"acronym"`
	MicCode        string `json:"mic_code"`
	Polity         string `json:"polity"`
	Currency       string `json:"currency"`
	TimeZone       int    `json:"time_zone"`
	OpenTime       string `json:"open_time"`
	CloseTime      string `json:"close_time"`
	TradingEnabled bool   `json:"trading_enabled"`
}

func ToExchangeResponse(e model.Exchange) ExchangeResponse {
	return ExchangeResponse{
		ExchangeID:     e.ExchangeID,
		Name:           e.Name,
		Acronym:        e.Acronym,
		MicCode:        e.MicCode,
		Polity:         e.Polity,
		Currency:       e.Currency,
		TimeZone:       e.TimeZone,
		OpenTime:       e.OpenTime,
		CloseTime:      e.CloseTime,
		TradingEnabled: e.TradingEnabled,
	}
}

func ToExchangeResponseList(exchanges []model.Exchange) []ExchangeResponse {
	result := make([]ExchangeResponse, len(exchanges))
	for i, e := range exchanges {
		result[i] = ToExchangeResponse(e)
	}
	return result
}

package repository

import "github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"

type OptionRepository interface {
	Upsert(option *model.Option) error
}

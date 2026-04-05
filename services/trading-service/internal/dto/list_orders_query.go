package dto

import "github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"

type ListOrdersQuery struct {
	Page      int                   `form:"page"`
	PageSize  int                   `form:"page_size"`
	Status    *model.OrderStatus    `form:"status"`
	Direction *model.OrderDirection `form:"direction"`
	IsDone    *bool                 `form:"is_done"`
}

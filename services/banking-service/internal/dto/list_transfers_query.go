package dto

type ListTransfersQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

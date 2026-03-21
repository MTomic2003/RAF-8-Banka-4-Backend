package client

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
)

type UserClient interface {
	GetClientByID(ctx context.Context, id uint) (*pb.GetClientByIdResponse, error)
	GetEmployeeByID(ctx context.Context, id uint) (*pb.GetEmployeeByIdResponse, error)
}

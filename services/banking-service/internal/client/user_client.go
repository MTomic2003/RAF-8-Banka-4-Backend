package client

import (
	"common/pkg/pb"
	"context"
)

type UserClient interface {
	GetClientByID(ctx context.Context, id uint) (*pb.GetClientByIdResponse, error)
	GetEmployeeByID(ctx context.Context, id uint) (*pb.GetEmployeeByIdResponse, error)
}

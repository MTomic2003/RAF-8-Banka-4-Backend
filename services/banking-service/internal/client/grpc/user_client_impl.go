package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
)

type UserServiceClient struct {
	client pb.UserServiceClient
}

func NewUserServiceClient(conn *grpc.ClientConn) *UserServiceClient {
	return &UserServiceClient{client: pb.NewUserServiceClient(conn)}
}

func (c *UserServiceClient) GetClientByID(ctx context.Context, id uint) (*pb.GetClientByIdResponse, error) {
	return c.client.GetClientById(ctx, &pb.GetClientByIdRequest{Id: uint64(id)})
}

func (c *UserServiceClient) GetEmployeeByID(ctx context.Context, id uint) (*pb.GetEmployeeByIdResponse, error) {
	return c.client.GetEmployeeById(ctx, &pb.GetEmployeeByIdRequest{Id: uint64(id)})
}

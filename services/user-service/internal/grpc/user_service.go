package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/repository"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	clientRepo   repository.ClientRepository
	employeeRepo repository.EmployeeRepository
}

func NewUserService(clientRepo repository.ClientRepository, employeeRepo repository.EmployeeRepository) *UserService {
	return &UserService{clientRepo: clientRepo, employeeRepo: employeeRepo}
}

func (s *UserService) GetClientById(ctx context.Context, req *pb.GetClientByIdRequest) (*pb.GetClientByIdResponse, error) {
	client, err := s.clientRepo.FindByID(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch client: %v", err)
	}
	if client == nil {
		return nil, status.Errorf(codes.NotFound, "client %d not found", req.Id)
	}
	return &pb.GetClientByIdResponse{
		Id:       uint64(client.ClientID),
		Email:    client.Identity.Email,
		FullName: client.FirstName + " " + client.LastName,
	}, nil
}

func (s *UserService) GetEmployeeById(ctx context.Context, req *pb.GetEmployeeByIdRequest) (*pb.GetEmployeeByIdResponse, error) {
	employee, err := s.employeeRepo.FindByID(ctx, uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch employee: %v", err)
	}
	if employee == nil {
		return nil, status.Errorf(codes.NotFound, "employee %d not found", req.Id)
	}
	resp := &pb.GetEmployeeByIdResponse{
		Id:           uint64(employee.EmployeeID),
		Email:        employee.Identity.Email,
		FullName:     employee.FirstName + " " + employee.LastName,
		IsSupervisor: employee.IsSupervisor(),
		IsAgent:      employee.IsAgent(),
	}

	if employee.ActuaryInfo != nil {
		resp.NeedApproval = employee.ActuaryInfo.NeedApproval
		resp.OrderLimit = employee.ActuaryInfo.Limit
		resp.UsedLimit = employee.ActuaryInfo.UsedLimit
	}

	return resp, nil
}

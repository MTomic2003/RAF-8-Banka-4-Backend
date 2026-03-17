package grpc

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/jwt"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
)

type PermissionService struct {
	pb.UnimplementedPermissionServiceServer
	provider auth.PermissionProvider
}

func NewPermissionService(provider auth.PermissionProvider) *PermissionService {
	return &PermissionService{provider: provider}
}

func (s *PermissionService) GetPermissions(ctx context.Context, req *pb.GetPermissionsRequest) (*pb.GetPermissionsResponse, error) {
	// TODO: Here we also return empty response if its not employee, should we considered in the future
	if auth.IdentityType(req.GetIdentityType()) != auth.IdentityEmployee {
		return &pb.GetPermissionsResponse{Permissions: []string{}}, nil
	}

	claims := &jwt.Claims{
		IdentityID:   uint(req.GetIdentityId()),
		IdentityType: req.GetIdentityType(),
		EmployeeID:   new(uint(req.GetSubjectId())),
	}

	permissions, err := s.provider.GetPermissions(ctx, claims)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(permissions))
	for _, p := range permissions {
		result = append(result, string(p))
	}

	return &pb.GetPermissionsResponse{Permissions: result}, nil
}

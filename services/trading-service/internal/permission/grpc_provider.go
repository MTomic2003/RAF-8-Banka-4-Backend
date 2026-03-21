package permission

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/jwt"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	perm "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
)

type GrpcPermissionProvider struct {
	client pb.PermissionServiceClient
}

func NewGrpcPermissionProvider(client pb.PermissionServiceClient) *GrpcPermissionProvider {
	return &GrpcPermissionProvider{client: client}
}

func (p *GrpcPermissionProvider) GetPermissions(ctx context.Context, claims *jwt.Claims) ([]perm.Permission, error) {
	if auth.IdentityType(claims.IdentityType) != auth.IdentityEmployee {
		return []perm.Permission{}, nil
	}

	req := &pb.GetPermissionsRequest{
		IdentityId:   uint64(claims.IdentityID),
		IdentityType: claims.IdentityType,
		SubjectId:    uint64(*claims.EmployeeID),
	}

	resp, err := p.client.GetPermissions(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]perm.Permission, 0, len(resp.GetPermissions()))
	for _, permName := range resp.GetPermissions() {
		result = append(result, perm.Permission(permName))
	}

	return result, nil
}

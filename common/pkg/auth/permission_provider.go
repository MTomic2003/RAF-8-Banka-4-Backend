package auth

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/jwt"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
)

// PermissionProvider loads all permissions for an identity. Used by
// Middleware to populate AuthContext.Permissions on every
// authenticated request.
//
// Implementations:
//   - DBPermissionProvider:   queries the DB directly (used by user-service)
//   - GRPCPermissionProvider: calls user-service over gRPC (used by other services)
type PermissionProvider interface {
	GetPermissions(ctx context.Context, claims *jwt.Claims) ([]permission.Permission, error)
}

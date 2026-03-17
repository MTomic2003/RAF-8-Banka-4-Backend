package permission

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	commonjwt "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/jwt"
	perm "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
)

// DBPermissionProvider loads all permissions for an identity by querying the
// database directly.
type DBPermissionProvider struct {
	db *gorm.DB
}

func NewDBPermissionProvider(db *gorm.DB) *DBPermissionProvider {
	return &DBPermissionProvider{db: db}
}

func (p *DBPermissionProvider) GetPermissions(ctx context.Context, claims *commonjwt.Claims) ([]perm.Permission, error) {
	if auth.IdentityType(claims.IdentityType) != auth.IdentityEmployee {
		return []perm.Permission{}, nil
	}

	employeeID := claims.EmployeeID

	if employeeID == nil {
		var resolvedEmployeeID uint

		err := p.db.WithContext(ctx).
			Table("employees").
			Select("employee_id").
			Where("identity_id = ?", claims.IdentityID).
			Scan(&resolvedEmployeeID).Error

		if err != nil {
			return nil, err
		}

		employeeID = &resolvedEmployeeID
	}

	var permissions []perm.Permission
	err := p.db.WithContext(ctx).
		Table("employee_permissions").
		Where("employee_id = ?", *employeeID).
		Pluck("permission", &permissions).Error

	return permissions, err
}

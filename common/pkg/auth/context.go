package auth

import (
	"common/pkg/errors"
	"common/pkg/permission"
	"context"

	"github.com/gin-gonic/gin"
)

// AuthContext holds the authenticated identity and permissions.
// Populated by Middleware, then available to RequirePermission and
// all downstream handlers via GetAuth / GetAuthFromContext.
type AuthContext struct {
	IdentityID   uint
	IdentityType IdentityType
	ClientID     *uint
	EmployeeID   *uint
	Permissions  []permission.Permission
}

type authKeyType struct{}

var authKey = authKeyType{}

// SetAuth stores the authenticated identity in both the Gin context and the
// request's stdlib context. The Gin context is used by middleware (GetAuth),
// and the stdlib context is used by service-layer code (GetAuthFromContext).
func SetAuth(c *gin.Context, auth *AuthContext) {
	c.Set(authKey, auth)
	ctx := context.WithValue(c.Request.Context(), authKey, auth)
	c.Request = c.Request.WithContext(ctx)
}

// SetAuthOnContext stores AuthContext into a stdlib context.
// This is intended for use in service-layer unit tests where
// there is no Gin context available.
func SetAuthOnContext(ctx context.Context, ac *AuthContext) context.Context {
	return context.WithValue(ctx, authKey, ac)
}

// GetAuth retrieves the authenticated identity from the Gin context.
func GetAuth(c *gin.Context) *AuthContext {
	val, exists := c.Get(authKey)
	if !exists {
		return nil
	}

	auth, ok := val.(*AuthContext)
	if !ok {
		return nil
	}

	return auth
}

// GetAuthFromContext retrieves the authenticated identity from a stdlib context.
func GetAuthFromContext(ctx context.Context) *AuthContext {
	val := ctx.Value(authKey)
	if val == nil {
		return nil
	}

	auth, ok := val.(*AuthContext)
	if !ok {
		return nil
	}

	return auth
}

// GetSubjectFromContext resolves the authenticated subject id from stdlib context.
// For client identities this is ClientID, and for employee identities this is EmployeeID.
func GetSubjectFromContext(ctx context.Context) (uint, error) {
	authCtx := GetAuthFromContext(ctx)
	if authCtx == nil {
		return 0, errors.UnauthorizedErr("not authenticated")
	}

	if authCtx.IdentityType == IdentityClient {
		if authCtx.ClientID == nil {
			return 0, errors.UnauthorizedErr("not authenticated")
		}
		
		return *authCtx.ClientID, nil
	}

	if authCtx.IdentityType == IdentityEmployee {
		if authCtx.EmployeeID == nil {
			return 0, errors.UnauthorizedErr("not authenticated")
		}
		return *authCtx.EmployeeID, nil
	}

	return 0, errors.ForbiddenErr("access denied for this identity type")
}

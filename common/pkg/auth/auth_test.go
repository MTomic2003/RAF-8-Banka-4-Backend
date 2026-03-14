package auth_test

import (
	"common/pkg/auth"
	"common/pkg/errors"
	"common/pkg/jwt"
	"common/pkg/logging"
	"common/pkg/permission"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeTokenVerifier struct {
	claims *jwt.Claims
	err    error
}

func (f *fakeTokenVerifier) VerifyToken(_ string) (*jwt.Claims, error) {
	return f.claims, f.err
}

type fakePermissionProvider struct {
	perms []permission.Permission
	err   error
}

func (f *fakePermissionProvider) GetPermissions(_ context.Context, _ uint) ([]permission.Permission, error) {
	return f.perms, f.err
}

func init() {
	gin.SetMode(gin.TestMode)
	logging.Init("test")
}

func middlewareRouter(verifier auth.TokenVerifier, provider auth.PermissionProvider) *gin.Engine {
	router := gin.New()
	router.Use(errors.ErrorHandler())
	router.Use(auth.Middleware(verifier, provider))
	router.GET("/test", func(c *gin.Context) {
		ac := auth.GetAuth(c)
		c.JSON(http.StatusOK, gin.H{"user_id": ac.UserID})
	})
	return router
}

func doGet(router *gin.Engine, authHeader string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	router.ServeHTTP(w, req)
	return w
}

func TestMiddleware_ValidToken(t *testing.T) {
	t.Parallel()

	router := middlewareRouter(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 42}},
		&fakePermissionProvider{perms: []permission.Permission{permission.EmployeeView}},
	)

	w := doGet(router, "Bearer valid-token")
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "42")
}

func TestMiddleware_MissingHeader(t *testing.T) {
	t.Parallel()

	router := middlewareRouter(
		&fakeTokenVerifier{},
		&fakePermissionProvider{},
	)

	w := doGet(router, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_MalformedHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "Token abc123"},
		{"basic auth", "Basic abc123"},
		{"only token value", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			router := middlewareRouter(
				&fakeTokenVerifier{claims: &jwt.Claims{UserID: 1}},
				&fakePermissionProvider{perms: []permission.Permission{}},
			)

			w := doGet(router, tt.header)
			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	t.Parallel()

	router := middlewareRouter(
		&fakeTokenVerifier{err: fmt.Errorf("invalid token")},
		&fakePermissionProvider{},
	)

	w := doGet(router, "Bearer bad-token")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_PermissionProviderError(t *testing.T) {
	t.Parallel()

	router := middlewareRouter(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 1}},
		&fakePermissionProvider{err: fmt.Errorf("db down")},
	)

	w := doGet(router, "Bearer valid-token")
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMiddleware_SetsAuthContext(t *testing.T) {
	t.Parallel()

	var captured *auth.AuthContext
	router := gin.New()
	router.Use(auth.Middleware(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 7}},
		&fakePermissionProvider{perms: []permission.Permission{permission.EmployeeView, permission.EmployeeCreate}},
	))
	router.GET("/test", func(c *gin.Context) {
		captured = auth.GetAuth(c)
		c.Status(http.StatusOK)
	})

	doGet(router, "Bearer valid-token")

	require.NotNil(t, captured)
	require.Equal(t, uint(7), captured.UserID)
	require.ElementsMatch(t, []permission.Permission{permission.EmployeeView, permission.EmployeeCreate}, captured.Permissions)
}

func TestMiddleware_SetsStdlibContext(t *testing.T) {
	t.Parallel()

	var captured *auth.AuthContext
	router := gin.New()
	router.Use(auth.Middleware(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 7}},
		&fakePermissionProvider{perms: []permission.Permission{permission.EmployeeView}},
	))
	router.GET("/test", func(c *gin.Context) {
		captured = auth.GetAuthFromContext(c.Request.Context())
		c.Status(http.StatusOK)
	})

	doGet(router, "Bearer valid-token")

	require.NotNil(t, captured)
	require.Equal(t, uint(7), captured.UserID)
}

func requirePermissionRouter(verifier auth.TokenVerifier, provider auth.PermissionProvider, required ...permission.Permission) *gin.Engine {
	router := gin.New()
	router.Use(errors.ErrorHandler())
	router.Use(auth.Middleware(verifier, provider))
	router.GET("/test", auth.RequirePermission(required...), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func TestRequirePermission_HasPermission(t *testing.T) {
	t.Parallel()

	router := requirePermissionRouter(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 1}},
		&fakePermissionProvider{perms: []permission.Permission{permission.EmployeeView, permission.EmployeeCreate}},
		permission.EmployeeView,
	)

	w := doGet(router, "Bearer valid-token")
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_HasMultipleRequired(t *testing.T) {
	t.Parallel()

	router := requirePermissionRouter(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 1}},
		&fakePermissionProvider{perms: []permission.Permission{permission.EmployeeView, permission.EmployeeCreate, permission.EmployeeUpdate}},
		permission.EmployeeView, permission.EmployeeCreate,
	)

	w := doGet(router, "Bearer valid-token")
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_MissingPermission(t *testing.T) {
	t.Parallel()

	router := requirePermissionRouter(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 1}},
		&fakePermissionProvider{perms: []permission.Permission{permission.EmployeeView}},
		permission.EmployeeDelete,
	)

	w := doGet(router, "Bearer valid-token")
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequirePermission_NoAuthContext(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(errors.ErrorHandler())
	router.GET("/test", auth.RequirePermission(permission.EmployeeView), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := doGet(router, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_EmptyUserPermissions(t *testing.T) {
	t.Parallel()

	router := requirePermissionRouter(
		&fakeTokenVerifier{claims: &jwt.Claims{UserID: 1}},
		&fakePermissionProvider{perms: []permission.Permission{}},
		permission.EmployeeView,
	)

	w := doGet(router, "Bearer valid-token")
	require.Equal(t, http.StatusForbidden, w.Code)
}

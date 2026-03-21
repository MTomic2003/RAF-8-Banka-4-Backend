package jwt_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/jwt"
)

func TestGenerateAndVerify(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	token, err := jwt.GenerateToken(&jwt.Claims{
		IdentityID:   42,
		IdentityType: "employee",
		EmployeeID:   uintPtr(8),
	}, secret, 15)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	verifier := jwt.NewJWTVerifier(secret)
	claims, err := verifier.VerifyToken(token)
	require.NoError(t, err)
	require.Equal(t, uint(42), claims.IdentityID)
	require.Equal(t, "employee", claims.IdentityType)
	require.NotNil(t, claims.EmployeeID)
	require.Equal(t, uint(8), *claims.EmployeeID)
	require.Nil(t, claims.ClientID)
}

func TestVerify_WrongSecret(t *testing.T) {
	t.Parallel()

	token, err := jwt.GenerateToken(&jwt.Claims{
		IdentityID:   42,
		IdentityType: "employee",
		EmployeeID:   uintPtr(8),
	}, "secret-a", 15)
	require.NoError(t, err)

	verifier := jwt.NewJWTVerifier("secret-b")
	_, err = verifier.VerifyToken(token)
	require.Error(t, err)
}

func TestVerify_TamperedToken(t *testing.T) {
	t.Parallel()

	token, err := jwt.GenerateToken(&jwt.Claims{
		IdentityID:   42,
		IdentityType: "employee",
		EmployeeID:   uintPtr(8),
	}, "test-secret", 15)
	require.NoError(t, err)

	verifier := jwt.NewJWTVerifier("test-secret")
	_, err = verifier.VerifyToken(token + "tampered")
	require.Error(t, err)
}

func TestVerify_ExpiredToken(t *testing.T) {
	t.Parallel()

	token, err := jwt.GenerateToken(&jwt.Claims{
		IdentityID:   42,
		IdentityType: "employee",
		EmployeeID:   uintPtr(8),
	}, "test-secret", -1)
	require.NoError(t, err)

	verifier := jwt.NewJWTVerifier("test-secret")
	_, err = verifier.VerifyToken(token)
	require.Error(t, err)
}

func TestVerify_EmptyToken(t *testing.T) {
	t.Parallel()

	verifier := jwt.NewJWTVerifier("test-secret")
	_, err := verifier.VerifyToken("")
	require.Error(t, err)
}

func TestVerify_GarbageToken(t *testing.T) {
	t.Parallel()

	verifier := jwt.NewJWTVerifier("test-secret")
	_, err := verifier.VerifyToken("not.a.jwt")
	require.Error(t, err)
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	verifier := jwt.NewJWTVerifier(secret)

	token1, err := jwt.GenerateToken(&jwt.Claims{
		IdentityID:   1,
		IdentityType: "employee",
		EmployeeID:   uintPtr(8),
	}, secret, 15)
	require.NoError(t, err)

	token2, err := jwt.GenerateToken(&jwt.Claims{
		IdentityID:   2,
		IdentityType: "client",
		ClientID:     uintPtr(17),
	}, secret, 15)
	require.NoError(t, err)

	require.NotEqual(t, token1, token2)

	claims1, err := verifier.VerifyToken(token1)
	require.NoError(t, err)
	require.Equal(t, uint(1), claims1.IdentityID)
	require.Equal(t, "employee", claims1.IdentityType)
	require.NotNil(t, claims1.EmployeeID)
	require.Equal(t, uint(8), *claims1.EmployeeID)
	require.Nil(t, claims1.ClientID)

	claims2, err := verifier.VerifyToken(token2)
	require.NoError(t, err)
	require.Equal(t, uint(2), claims2.IdentityID)
	require.Equal(t, "client", claims2.IdentityType)
	require.NotNil(t, claims2.ClientID)
	require.Equal(t, uint(17), *claims2.ClientID)
	require.Nil(t, claims2.EmployeeID)
}

func uintPtr(v uint) *uint {
	return &v
}

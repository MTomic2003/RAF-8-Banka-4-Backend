package jwt_test

import (
	"common/pkg/jwt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateAndVerify(t *testing.T) {
	t.Parallel()

	secret := "test-secret"
	token, err := jwt.GenerateToken(42, secret, 15)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	verifier := jwt.NewJWTVerifier(secret)
	claims, err := verifier.VerifyToken(token)
	require.NoError(t, err)
	require.Equal(t, uint(42), claims.UserID)
}

func TestVerify_WrongSecret(t *testing.T) {
	t.Parallel()

	token, err := jwt.GenerateToken(42, "secret-a", 15)
	require.NoError(t, err)

	verifier := jwt.NewJWTVerifier("secret-b")
	_, err = verifier.VerifyToken(token)
	require.Error(t, err)
}

func TestVerify_TamperedToken(t *testing.T) {
	t.Parallel()

	token, err := jwt.GenerateToken(42, "test-secret", 15)
	require.NoError(t, err)

	verifier := jwt.NewJWTVerifier("test-secret")
	_, err = verifier.VerifyToken(token + "tampered")
	require.Error(t, err)
}

func TestVerify_ExpiredToken(t *testing.T) {
	t.Parallel()

	token, err := jwt.GenerateToken(42, "test-secret", -1)
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

	token1, err := jwt.GenerateToken(1, secret, 15)
	require.NoError(t, err)

	token2, err := jwt.GenerateToken(2, secret, 15)
	require.NoError(t, err)

	require.NotEqual(t, token1, token2)

	claims1, err := verifier.VerifyToken(token1)
	require.NoError(t, err)
	require.Equal(t, uint(1), claims1.UserID)

	claims2, err := verifier.VerifyToken(token2)
	require.NoError(t, err)
	require.Equal(t, uint(2), claims2.UserID)
}

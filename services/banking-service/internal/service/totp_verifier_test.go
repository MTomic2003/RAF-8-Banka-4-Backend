package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVerifyTOTPCode(t *testing.T) {
	t.Parallel()

	now := time.Unix(59, 0)

	require.True(t, verifyTOTPCode("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", "287082", now, 0))
	require.True(t, verifyTOTPCode("jbswy3dpehpk3pxp", "282760", time.Unix(0, 0), 0))
	require.False(t, verifyTOTPCode("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", "123456", now, 0))
	require.False(t, verifyTOTPCode("", "123456", now, 0))
	require.False(t, verifyTOTPCode("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", "12345", now, 0))
}

func TestVerifyTOTPCode_AllowedSkew(t *testing.T) {
	t.Parallel()

	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	now := time.Unix(61, 0)

	require.True(t, verifyTOTPCode(secret, "287082", now, 1))
	require.False(t, verifyTOTPCode(secret, "287082", now, 0))
}

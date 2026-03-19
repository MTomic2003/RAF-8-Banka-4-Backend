package service

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	totpDigits      = 6
	totpStepSeconds = 30
	totpAllowedSkew = 10
)

func verifyTOTPCode(secretBase32, code string, now time.Time, allowedSkew int) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}

	secret, err := decodeBase32Secret(secretBase32)
	if err != nil {
		return false
	}

	counter := now.Unix() / totpStepSeconds
	for offset := -allowedSkew; offset <= allowedSkew; offset++ {
		expected := generateTOTP(secret, counter+int64(offset))
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}

	return false
}

func decodeBase32Secret(secretBase32 string) ([]byte, error) {
	normalized := strings.ToUpper(strings.TrimSpace(secretBase32))
	normalized = strings.ReplaceAll(normalized, " ", "")
	if normalized == "" {
		return nil, fmt.Errorf("empty secret")
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalized)
}

func generateTOTP(secret []byte, counter int64) string {
	var message [8]byte
	binary.BigEndian.PutUint64(message[:], uint64(counter))

	h := hmac.New(sha1.New, secret)
	_, _ = h.Write(message[:])
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f

	binaryCode := (int(hash[offset])&0x7f)<<24 |
		(int(hash[offset+1])&0xff)<<16 |
		(int(hash[offset+2])&0xff)<<8 |
		(int(hash[offset+3]) & 0xff)
	otp := binaryCode % 1000000

	return fmt.Sprintf("%06d", otp)
}

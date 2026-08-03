package partner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign returns the lowercase hex HMAC-SHA256 of payload keyed by secret. This is the exact
// scheme the Peccancy platform validates against; the client uses it for you.
func Sign(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// safeEqual is a constant-time comparison of two hex signatures.
func safeEqual(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

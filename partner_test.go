package partner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func TestSignMatchesHMAC(t *testing.T) {
	h := hmac.New(sha256.New, []byte("key"))
	h.Write([]byte("abc"))
	want := hex.EncodeToString(h.Sum(nil))
	if got := Sign("abc", "key"); got != want {
		t.Fatalf("Sign mismatch: got %s want %s", got, want)
	}
}

func TestVerifyCallbackRoundTrip(t *testing.T) {
	const secret = "s3cr3t"
	ts := time.Now().Unix()

	cb := Callback{TransactionID: "tx-1", Status: "confirmed", Amount: 9.99, Timestamp: ts}
	cb.Signature = Sign(fmt.Sprintf("%s:%s:%s:%d", cb.TransactionID, cb.Status, money(cb.Amount), ts), secret)
	if !VerifyCallback(cb, secret, 120*time.Second) {
		t.Fatal("valid callback should verify")
	}

	tampered := cb
	tampered.Amount = 10.0
	if VerifyCallback(tampered, secret, 120*time.Second) {
		t.Fatal("tampered callback must fail")
	}

	stale := Callback{TransactionID: "tx-1", Status: "confirmed", Amount: 9.99, Timestamp: ts - 1000}
	stale.Signature = Sign(fmt.Sprintf("%s:%s:%s:%d", stale.TransactionID, stale.Status, money(stale.Amount), stale.Timestamp), secret)
	if VerifyCallback(stale, secret, 120*time.Second) {
		t.Fatal("stale callback must fail")
	}
}

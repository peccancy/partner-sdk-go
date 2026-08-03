// Webhook receiver example: verify signed callbacks from the platform.
//
// Point your partner's callback_url at http(s)://your-host/peccancy/callback
//
// Run: PECCANCY_PARTNER_SECRET=... go run ./examples/webhook-receiver
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	partner "github.com/peccancy/partner-sdk-go"
)

func main() {
	secret := os.Getenv("PECCANCY_PARTNER_SECRET")
	if secret == "" {
		log.Fatal("missing env PECCANCY_PARTNER_SECRET")
	}

	http.HandleFunc("/peccancy/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var cb partner.Callback
		if err := json.NewDecoder(r.Body).Decode(&cb); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if !partner.VerifyCallback(cb, secret, 120*time.Second) {
			http.Error(w, "bad signature", http.StatusUnauthorized)
			return
		}

		// Verified & fresh — safe to act on.
		log.Println("callback:", cb.TransactionID, cb.Status, cb.Amount)
		// ... credit the user / mark the order paid ...

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Println("listening on :3000/peccancy/callback")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

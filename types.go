package partner

import "time"

// Variant is one outcome option of a dispute.
type Variant struct {
	Description string `json:"description"`
}

// CreateDisputeInput is the input to CreateDispute.
type CreateDisputeInput struct {
	Description string    // dispute question / title
	Variants    []Variant // 2–100 outcomes
	FinishDate  time.Time // when the event ends
	StopDate    time.Time // when betting closes (before FinishDate)
	MinBet      *int64    // optional
	MaxBet      *int64    // optional
	Lang        string    // optional (default "en")
	IsClosed    bool      // if true, only you resolve the winner
}

// DisputeVariant is a variant as returned by the API.
type DisputeVariant struct {
	ID          string
	Description string
	CountOfBets int
	Amount      float64
}

// Dispute is a created dispute.
type Dispute struct {
	ID          string
	Description string
	Variants    []DisputeVariant
	TeamID      string
	Status      string
	TotalAmount float64
	FinishDate  string
	StopDate    string
	CreatedAt   int64
}

// UserIdentifier identifies a user for payments.
type UserIdentifier struct {
	Type  string // "email" | "phone" | "id"
	Value string
}

// InitPaymentResult is the result of InitPayment.
type InitPaymentResult struct {
	TransactionID string
	Status        string
	Amount        float64
}

// InvoiceResult is the result of CreateInvoice.
type InvoiceResult struct {
	InvoiceID  string
	InvoiceURL string
	Status     string
	Amount     float64
}

// Callback is the signed JSON body the platform POSTs to your callback_url.
type Callback struct {
	TransactionID string  `json:"transaction_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	Timestamp     int64   `json:"timestamp"`
	Signature     string  `json:"signature"`
}

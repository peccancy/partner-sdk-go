// Package partner is the official Peccancy partner SDK: create disputes, take payments and
// verify signed result webhooks. Every request is authenticated for you with HMAC-SHA256.
//
//	c, _ := partner.New("https://disputes.online/partner", partnerID, secret)
//	d, err := c.CreateDispute(ctx, partner.CreateDisputeInput{ ... })
package partner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client is the entry point of the SDK. Construct once with New and reuse.
type Client struct {
	baseURL    string
	partnerID  string
	secret     string
	httpClient *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithTimeout sets the request timeout (default 15s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient = &http.Client{Timeout: d} }
}

// New creates a Client. baseURL is the partner API base, e.g. "https://disputes.online/partner".
func New(baseURL, partnerID, secret string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}
	if partnerID == "" {
		return nil, fmt.Errorf("partnerID is required")
	}
	if secret == "" {
		return nil, fmt.Errorf("secret is required")
	}
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		partnerID:  partnerID,
		secret:     secret,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func nowSec() int64          { return time.Now().Unix() }
func money(n float64) string { return fmt.Sprintf("%.2f", n) }

// ---- Disputes ----

// CreateDispute creates a dispute (question + 2–100 outcomes).
func (c *Client) CreateDispute(ctx context.Context, in CreateDisputeInput) (*Dispute, error) {
	ts := nowSec()
	payload := fmt.Sprintf("%s:%s:%d", c.partnerID, in.Description, ts)
	body := map[string]any{
		"description": in.Description,
		"variants":    in.Variants,
		"finish_date": in.FinishDate.Format(time.RFC3339),
		"stop_date":   in.StopDate.Format(time.RFC3339),
	}
	if in.MinBet != nil {
		body["min_bet"] = *in.MinBet
	}
	if in.MaxBet != nil {
		body["max_bet"] = *in.MaxBet
	}
	if in.Lang != "" {
		body["lang"] = in.Lang
	}
	if in.IsClosed {
		body["is_closed"] = true
	}

	var out struct {
		Data struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			Variants    []struct {
				ID          string  `json:"id"`
				Description string  `json:"description"`
				CountOfBets int     `json:"count_of_bets"`
				Amount      float64 `json:"amount"`
			} `json:"variants"`
			TeamID      string  `json:"team_id"`
			Status      string  `json:"status"`
			TotalAmount float64 `json:"total_amount"`
			FinishDate  string  `json:"finish_date"`
			StopDate    string  `json:"stop_date"`
			CreatedAt   int64   `json:"created_at"`
		} `json:"data"`
	}
	if err := c.headerSigned(ctx, http.MethodPost, "/api/v1/partner/disputes", payload, ts, body, &out); err != nil {
		return nil, err
	}

	d := &Dispute{
		ID:          out.Data.ID,
		Description: out.Data.Description,
		TeamID:      out.Data.TeamID,
		Status:      out.Data.Status,
		TotalAmount: out.Data.TotalAmount,
		FinishDate:  out.Data.FinishDate,
		StopDate:    out.Data.StopDate,
		CreatedAt:   out.Data.CreatedAt,
	}
	for _, v := range out.Data.Variants {
		d.Variants = append(d.Variants, DisputeVariant{ID: v.ID, Description: v.Description, CountOfBets: v.CountOfBets, Amount: v.Amount})
	}
	return d, nil
}

// StopBets closes betting on a dispute.
func (c *Client) StopBets(ctx context.Context, disputeID string) error {
	ts := nowSec()
	payload := fmt.Sprintf("%s:%s:%d", c.partnerID, disputeID, ts)
	return c.headerSigned(ctx, http.MethodPut, "/api/v1/partner/disputes/"+disputeID+"/stopbet", payload, ts, nil, nil)
}

// StartGame marks a dispute in-progress.
func (c *Client) StartGame(ctx context.Context, disputeID string) error {
	ts := nowSec()
	payload := fmt.Sprintf("%s:%s:%d", c.partnerID, disputeID, ts)
	return c.headerSigned(ctx, http.MethodPut, "/api/v1/partner/disputes/"+disputeID+"/startgame", payload, ts, nil, nil)
}

// SetWinner declares the winning outcome by its variant description (team name).
func (c *Client) SetWinner(ctx context.Context, disputeID, winnerTeamName string) error {
	ts := nowSec()
	payload := fmt.Sprintf("%s:%s:%s:%d", c.partnerID, disputeID, winnerTeamName, ts)
	return c.headerSigned(ctx, http.MethodPost, "/api/v1/partner/disputes/"+disputeID+"/winner", payload, ts,
		map[string]string{"winner_team_name": winnerTeamName}, nil)
}

// ---- Payments ----

// InitPayment charges a known user (by email/phone/id).
func (c *Client) InitPayment(ctx context.Context, amount float64, user UserIdentifier, description string) (*InitPaymentResult, error) {
	ts := nowSec()
	payload := fmt.Sprintf("%s:%s:%s:%d", c.partnerID, money(amount), user.Value, ts)
	body := map[string]any{
		"id":              c.partnerID,
		"amount":          amount,
		"user_identifier": map[string]string{"type": user.Type, "value": user.Value},
		"description":     description,
		"signature":       Sign(payload, c.secret),
		"timestamp":       ts,
	}
	var out struct {
		TransactionID string  `json:"transaction_id"`
		Status        string  `json:"status"`
		Amount        float64 `json:"amount"`
	}
	if err := c.bodySigned(ctx, "/partner/api/v1/init", body, &out); err != nil {
		return nil, err
	}
	return &InitPaymentResult{TransactionID: out.TransactionID, Status: out.Status, Amount: out.Amount}, nil
}

// CreateInvoice creates a hosted payment link (invoice).
func (c *Client) CreateInvoice(ctx context.Context, amount float64, description string) (*InvoiceResult, error) {
	ts := nowSec()
	payload := fmt.Sprintf("%s:%s:%s:%d", c.partnerID, money(amount), description, ts)
	body := map[string]any{
		"amount":      amount,
		"description": description,
		"signature":   Sign(payload, c.secret),
		"timestamp":   ts,
	}
	var out struct {
		InvoiceID  string  `json:"invoice_id"`
		InvoiceURL string  `json:"invoice_url"`
		Status     string  `json:"status"`
		Amount     float64 `json:"amount"`
	}
	if err := c.bodySigned(ctx, "/partner/api/v1/invoice", body, &out); err != nil {
		return nil, err
	}
	return &InvoiceResult{InvoiceID: out.InvoiceID, InvoiceURL: out.InvoiceURL, Status: out.Status, Amount: out.Amount}, nil
}

// ---- Webhooks ----

// VerifyCallback reports whether a callback POSTed to your callback_url is authentic: the
// signature must be valid and the timestamp fresh (within maxSkew). Always verify first.
func VerifyCallback(cb Callback, secret string, maxSkew time.Duration) bool {
	payload := fmt.Sprintf("%s:%s:%s:%d", cb.TransactionID, cb.Status, money(cb.Amount), cb.Timestamp)
	if !safeEqual(Sign(payload, secret), cb.Signature) {
		return false
	}
	age := time.Now().Unix() - cb.Timestamp
	if age < 0 {
		age = -age
	}
	return age <= int64(maxSkew.Seconds())
}

// ---- internals ----

func (c *Client) headerSigned(ctx context.Context, method, path, payload string, ts int64, body, out any) error {
	return c.request(ctx, method, path, map[string]string{
		"X-Partner-ID":        c.partnerID,
		"X-Partner-Timestamp": strconv.FormatInt(ts, 10),
		"X-Partner-Signature": Sign(payload, c.secret),
	}, body, out)
}

func (c *Client) bodySigned(ctx context.Context, path string, body, out any) error {
	return c.request(ctx, http.MethodPost, path, map[string]string{"X-Partner-ID": c.partnerID}, body, out)
}

func (c *Client) request(ctx context.Context, method, path string, headers map[string]string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &APIError{StatusCode: 0, Body: "", message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(data),
			message:    fmt.Sprintf("Peccancy API %s %s -> %d", method, path, resp.StatusCode),
		}
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return err
		}
	}
	return nil
}

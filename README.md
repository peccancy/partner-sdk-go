# partner-sdk-go

[![CI](https://github.com/peccancy/partner-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/peccancy/partner-sdk-go/actions/workflows/ci.yml)
[![Partner API docs](https://img.shields.io/badge/docs-Partner%20API-blue)](https://docs.disputes.online/swagger/index.html)
[![Go Reference](https://pkg.go.dev/badge/github.com/peccancy/partner-sdk-go.svg)](https://pkg.go.dev/github.com/peccancy/partner-sdk-go)

Official Go SDK for the **Peccancy** disputes/betting platform.

Connect your game or app once and let your users bet on outcomes: create disputes, control
their lifecycle, declare winners, take payments, and verify signed result webhooks. Every
request is authenticated for you with HMAC-SHA256 — you never build a signature by hand.

- Standard library only (no dependencies), Go 1.21+
- Same surface as our Node / PHP / Python SDKs

## Install

```bash
go get github.com/peccancy/partner-sdk-go
```

```go
import partner "github.com/peccancy/partner-sdk-go"
```

## Get credentials

1. Register as a partner at **https://disputes.online/profile?tab=partners** and open your partner.
2. Copy your **partnerID** (UUID) and **secret**.
3. Set a **callback_url** on your partner if you want result/payment webhooks.

## Quickstart

```go
c, err := partner.New("https://disputes.online/partner", partnerID, secret)
if err != nil { log.Fatal(err) }

now := time.Now()
d, err := c.CreateDispute(ctx, partner.CreateDisputeInput{
    Description: "Who wins Round 5?",
    Variants:    []partner.Variant{{Description: "Alice"}, {Description: "Bob"}},
    StopDate:    now.Add(5 * time.Minute),
    FinishDate:  now.Add(30 * time.Minute),
})
fmt.Println(d.ID)
```

## Disputes

```go
d, _ := c.CreateDispute(ctx, partner.CreateDisputeInput{
    Description: "Who wins?",
    Variants:    []partner.Variant{{Description: "Team A"}, {Description: "Team B"}},
    StopDate:    stop,
    FinishDate:  finish,
    Lang:        "en",   // optional
    // MinBet/MaxBet: *int64, IsClosed: bool (optional)
})

c.StopBets(ctx, d.ID)            // close betting
c.StartGame(ctx, d.ID)           // mark in-progress (optional)
c.SetWinner(ctx, d.ID, "Team A") // declare winner by variant description
```

## Payments

```go
tx, _ := c.InitPayment(ctx, 9.99, partner.UserIdentifier{Type: "email", Value: "user@example.com"}, "Coins pack")

inv, _ := c.CreateInvoice(ctx, 19.99, "Tournament entry")
fmt.Println(inv.InvoiceURL)
```

## Webhooks (results & payments)

The platform POSTs a signed JSON callback to your `callback_url`. **Always verify it**:

```go
var cb partner.Callback
_ = json.NewDecoder(r.Body).Decode(&cb)
if !partner.VerifyCallback(cb, secret, 120*time.Second) {
    http.Error(w, "bad signature", http.StatusUnauthorized)
    return
}
// ... credit the user / mark the order paid ...
```

`VerifyCallback` checks the HMAC signature **and** timestamp freshness.

## Authentication (under the hood)

The SDK adds `X-Partner-ID`, `X-Partner-Timestamp`, `X-Partner-Signature` headers (payment
endpoints put signature/timestamp in the body). The signature is a hex HMAC-SHA256 over a
colon-joined payload:

| Operation | Signed payload |
|-----------|----------------|
| CreateDispute | `partnerID:description:timestamp` |
| StopBets / StartGame | `partnerID:disputeID:timestamp` |
| SetWinner | `partnerID:disputeID:winnerTeamName:timestamp` |
| InitPayment | `partnerID:amount(2dp):userValue:timestamp` |
| CreateInvoice | `partnerID:amount(2dp):description:timestamp` |
| callback (inbound) | `transactionID:status:amount(2dp):timestamp` |

Keep your server clock in sync (NTP) — the platform rejects timestamps more than 120s off.

## Errors

Non-2xx responses return a `*partner.APIError` with `.StatusCode` and `.Body` (use `errors.As`).

## Examples

- [`examples/connect-your-game`](./examples/connect-your-game)
- [`examples/webhook-receiver`](./examples/webhook-receiver)

## Links

- **Register / get credentials:** https://disputes.online/profile?tab=partners
- **Partner API reference:** https://docs.disputes.online/swagger/index.html
- **Other SDKs:** [Node](https://github.com/peccancy/partner-sdk-node) · [PHP](https://github.com/peccancy/partner-sdk-php) · [Python](https://github.com/peccancy/partner-sdk-python) · [Go](https://github.com/peccancy/partner-sdk-go)

## License

MIT

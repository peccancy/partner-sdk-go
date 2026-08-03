// Connect-your-game example: the full dispute lifecycle for one match.
//
// Run: PECCANCY_PARTNER_ID=... PECCANCY_PARTNER_SECRET=... go run ./examples/connect-your-game
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	partner "github.com/peccancy/partner-sdk-go"
)

func main() {
	c, err := partner.New(
		envOr("PECCANCY_BASE_URL", "https://disputes.online/partner"),
		mustEnv("PECCANCY_PARTNER_ID"),
		mustEnv("PECCANCY_PARTNER_SECRET"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	now := time.Now()

	// 1) A new match starts -> open a dispute.
	dispute, err := c.CreateDispute(ctx, partner.CreateDisputeInput{
		Description: "Who wins the Alias round?",
		Variants:    []partner.Variant{{Description: "Red Team"}, {Description: "Blue Team"}},
		StopDate:    now.Add(2 * time.Minute),
		FinishDate:  now.Add(20 * time.Minute),
	})
	if err != nil {
		fail(err)
	}
	log.Println("created dispute:", dispute.ID)

	// 2) The round begins -> close betting and mark in-progress.
	if err := c.StopBets(ctx, dispute.ID); err != nil {
		fail(err)
	}
	if err := c.StartGame(ctx, dispute.ID); err != nil {
		fail(err)
	}
	log.Println("betting closed, game in progress")

	// 3) The round ends -> declare the winner by its name.
	if err := c.SetWinner(ctx, dispute.ID, "Red Team"); err != nil {
		fail(err)
	}
	log.Println("winner set: Red Team")
}

func fail(err error) {
	var apiErr *partner.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("API error %d: %s", apiErr.StatusCode, apiErr.Body)
	}
	log.Fatal(err)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing env %s", name)
	}
	return v
}

func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

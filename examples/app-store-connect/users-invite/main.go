// Example: invite a new team member with DEVELOPER access to a
// specific set of apps.
package main

import (
	"context"
	"errors"
	"log"
	"time"

	Apple "github.com/godrealms/go-apple-sdk"
	AppStoreConnect "github.com/godrealms/go-apple-sdk/app-store-connect"
)

func main() {
	kid := ""
	iss := ""
	bid := ""
	privateKey := ""

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Inspect the current team first — useful for idempotency checks
	// ("has this email already been invited or added?").
	list, err := svc.Users().List(ctx, AppStoreConnect.NewQuery().Limit(200))
	if err != nil {
		fatal("list users", err)
	}
	log.Printf("team has %d existing users", len(list.Data))

	fa := false
	inv, err := svc.UserInvitations().Create(ctx, AppStoreConnect.CreateUserInvitationRequest{
		Email:               "carol@example.com",
		FirstName:           "Carol",
		LastName:            "Cherry",
		Roles:               []string{"DEVELOPER"},
		AllAppsVisible:      &fa,
		ProvisioningAllowed: &fa,
		VisibleAppIDs:       []string{"1234567890"},
	})
	if err != nil {
		fatal("create invitation", err)
	}
	log.Printf("invitation %s sent to %s (expires %v)",
		inv.Id, inv.Attributes.Email, inv.Attributes.ExpirationDate)
}

func fatal(label string, err error) {
	var apiErr *AppStoreConnect.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("%s: Apple rejected the request: HTTP %d, %d errors", label, apiErr.StatusCode, len(apiErr.Errors))
	}
	log.Fatalf("%s: %v", label, err)
}

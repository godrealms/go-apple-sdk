// Example: create a non-subscription in-app purchase under an app,
// then create a companion auto-renewable subscription group.
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

	appID := "1234567890"

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// A. Non-consumable unlock (one-time purchase).
	familySharable := true
	iap, err := svc.InAppPurchases().Create(ctx, AppStoreConnect.CreateInAppPurchaseRequest{
		AppID:             appID,
		Name:              "Pro Unlock",
		ProductID:         "com.acme.widgets.prounlock",
		InAppPurchaseType: "NON_CONSUMABLE",
		ReviewNote:        "Unlocks the pro widget spinner permanently.",
		FamilySharable:    &familySharable,
	})
	if err != nil {
		fatal("create IAP", err)
	}
	log.Printf("IAP %s created (%s)", iap.Id, iap.Attributes.ProductID)

	// Rename without losing any other fields — the builder omits nil
	// attributes so Apple keeps the original reviewNote and sharing
	// flag untouched.
	if _, err := svc.InAppPurchases().Update(ctx, iap.Id, AppStoreConnect.NewInAppPurchaseUpdate().
		Name("Pro Widget Unlock")); err != nil {
		fatal("rename IAP", err)
	}

	// B. Subscription group container (for auto-renewing tiers).
	grp, err := svc.SubscriptionGroups().Create(ctx, AppStoreConnect.CreateSubscriptionGroupRequest{
		AppID:         appID,
		ReferenceName: "Acme Pro Plans",
	})
	if err != nil {
		fatal("create subscription group", err)
	}
	log.Printf("subscription group %s (%s) ready for tiers", grp.Id, grp.Attributes.ReferenceName)
}

func fatal(label string, err error) {
	var apiErr *AppStoreConnect.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("%s: Apple rejected the request: HTTP %d, %d errors", label, apiErr.StatusCode, len(apiErr.Errors))
	}
	log.Fatalf("%s: %v", label, err)
}

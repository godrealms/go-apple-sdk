// Example: fetch customer reviews for an app and post a developer
// response to the first 1-star review found.
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
	kid := ""        // Your private key ID
	iss := ""        // Your issuer ID
	bid := ""        // Your app's primary bundle ID
	privateKey := "" // Your ES256 private key (PKCS#8)
	appID := ""      // App Store Connect resource id of the target app

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Walk every page of reviews, newest-first.
	it := svc.CustomerReviews().ListForAppIterator(appID,
		AppStoreConnect.NewQuery().Sort("-createdDate").Limit(50))

	var firstBad *AppStoreConnect.CustomerReview
	var total int
	for it.Next(ctx) {
		page := it.Page()
		total += len(page.Data)
		for i, review := range page.Data {
			log.Printf("  [%d★] %s — %s", review.Attributes.Rating, review.Attributes.Title, review.Attributes.ReviewerNickname)
			if firstBad == nil && review.Attributes.Rating == 1 {
				firstBad = &page.Data[i]
			}
		}
	}
	if err := it.Err(); err != nil {
		var apiErr *AppStoreConnect.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("Apple rejected the request: HTTP %d", apiErr.StatusCode)
		}
		log.Fatalf("local error: %v", err)
	}
	log.Printf("scanned %d reviews", total)

	if firstBad == nil {
		log.Printf("no 1-star reviews to respond to")
		return
	}

	resp, err := svc.CustomerReviews().Respond(ctx, firstBad.Id,
		"Thanks for the feedback! Please email support@example.com so we can help.")
	if err != nil {
		log.Fatalf("respond: %v", err)
	}
	log.Printf("posted response %s (state=%s)", resp.Id, resp.Attributes.State)
}

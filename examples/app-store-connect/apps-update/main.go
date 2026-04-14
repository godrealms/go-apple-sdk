// Example: modify an app's primary locale and content-rights
// declaration via PATCH /v1/apps/{id}.
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
	kid := ""        // Your private key ID from App Store Connect
	iss := ""        // Your issuer ID
	bid := ""        // Your app's primary bundle ID
	privateKey := "" // Your ES256 private key (PKCS#8)
	appID := ""      // App Store Connect resource id of the app to modify

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	update := AppStoreConnect.NewAppUpdate().
		PrimaryLocale("en-US").
		AvailableInNewTerritories(true).
		ContentRightsDeclaration("DOES_NOT_USE_THIRD_PARTY_CONTENT")

	app, err := svc.Apps().Update(ctx, appID, update)
	if err != nil {
		var apiErr *AppStoreConnect.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("Apple rejected the update: HTTP %d, %d errors", apiErr.StatusCode, len(apiErr.Errors))
		}
		log.Fatalf("local error: %v", err)
	}

	log.Printf("updated app %s: primaryLocale=%s contentRights=%s",
		app.Id, app.Attributes.PrimaryLocale, app.Attributes.ContentRightsDeclaration)
}

// Example: list every App on your App Store Connect account using the
// new AppStoreConnect.Service + automatic pagination.
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
	kid := ""        // Your private key ID from App Store Connect (Ex: 2X9R4HXF34)
	iss := ""        // Your issuer ID from the Keys page in App Store Connect
	bid := ""        // Your app's primary bundle ID
	privateKey := "" // Your ES256 private key (PKCS#8)

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := AppStoreConnect.NewQuery().
		Fields("apps", "name", "bundleId", "sku", "primaryLocale").
		Sort("name").
		Limit(200)

	apps, err := svc.Apps().ListAll(ctx, query)
	if err != nil {
		var apiErr *AppStoreConnect.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("Apple rejected the request: HTTP %d, %d errors", apiErr.StatusCode, len(apiErr.Errors))
		}
		log.Fatalf("local error: %v", err)
	}

	log.Printf("account has %d apps", len(apps))
	for _, app := range apps {
		log.Printf("  %s  %s  (%s)", app.Id, app.Attributes.Name, app.Attributes.BundleId)
	}
}

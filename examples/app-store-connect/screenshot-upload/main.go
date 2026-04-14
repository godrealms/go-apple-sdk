// Example: upload a marketing screenshot to an App Store version
// localization using the three-step reserve → PUT → commit flow. The
// Upload helper wraps all three steps so this example only has to
// supply the raw image bytes.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	Apple "github.com/godrealms/go-apple-sdk"
	AppStoreConnect "github.com/godrealms/go-apple-sdk/app-store-connect"
)

func main() {
	kid := ""
	iss := ""
	bid := ""
	privateKey := ""

	// Replace with your actual localization id and PNG path. Apple
	// requires the filename embedded in the reservation to match the
	// bytes you upload, so pass a real filename.
	localizationID := "loc-123"
	imagePath := "./hero-6.7.png"

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 1. Ensure there's a screenshot set of the right display type.
	//    In a real pipeline you would list existing sets first to
	//    avoid duplicates.
	set, err := svc.AppScreenshotSets().Create(ctx, AppStoreConnect.CreateAppScreenshotSetRequest{
		LocalizationID:        localizationID,
		ScreenshotDisplayType: "APP_IPHONE_67",
	})
	if err != nil {
		fatal("create screenshot set", err)
	}
	log.Printf("screenshot set %s created (%s)", set.Id, set.Attributes.ScreenshotDisplayType)

	// 2. Read the PNG bytes and hand them to Upload. Upload reserves,
	//    executes every upload operation Apple hands back, computes
	//    the MD5, and commits in one call.
	data, err := os.ReadFile(imagePath)
	if err != nil {
		log.Fatalf("read image: %v", err)
	}
	shot, err := svc.AppScreenshots().Upload(ctx, set.Id, "hero-6.7.png", data)
	if err != nil {
		fatal("upload screenshot", err)
	}
	log.Printf("screenshot %s committed (%d bytes)", shot.Id, shot.Attributes.FileSize)
	if shot.Attributes.ImageAsset != nil {
		log.Printf("delivered asset: %s", shot.Attributes.ImageAsset.TemplateURL)
	}
}

func fatal(label string, err error) {
	var apiErr *AppStoreConnect.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("%s: Apple rejected the request: HTTP %d, %d errors", label, apiErr.StatusCode, len(apiErr.Errors))
	}
	log.Fatalf("%s: %v", label, err)
}

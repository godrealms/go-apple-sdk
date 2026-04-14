// Example: create an App Store version, attach a build, update the
// en-US release notes, then submit for review.
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

	// Replace with your actual resource ids.
	appID := "1234567890"
	buildID := "build-42"

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Create the version row.
	ver, err := svc.AppStoreVersions().Create(ctx, AppStoreConnect.CreateAppStoreVersionRequest{
		AppID:         appID,
		Platform:      "IOS",
		VersionString: "1.2.3",
		Copyright:     "© 2026 Acme",
		ReleaseType:   "AFTER_APPROVAL",
	})
	if err != nil {
		fatal("create version", err)
	}
	log.Printf("version %s created in state %s", ver.Id, ver.Attributes.AppStoreState)

	// 2. Bind the compiled build to this version.
	if err := svc.AppStoreVersions().SelectBuild(ctx, ver.Id, buildID); err != nil {
		fatal("select build", err)
	}

	// 3. Write the en-US release notes. In real use you would loop over
	//    every locale the app supports.
	loc, err := svc.AppStoreVersionLocalizations().Create(ctx, AppStoreConnect.CreateAppStoreVersionLocalizationRequest{
		VersionID:       ver.Id,
		Locale:          "en-US",
		WhatsNew:        "- Faster widget spinning\n- Dark mode polish\n- Misc. bug fixes",
		PromotionalText: "Our fastest spinner yet!",
	})
	if err != nil {
		fatal("create localization", err)
	}
	log.Printf("localization %s created (%s)", loc.Id, loc.Attributes.Locale)

	// 4. Submit for review — programmatic equivalent of the
	//    "Submit for Review" button in App Store Connect.
	sub, err := svc.AppStoreVersionSubmissions().Submit(ctx, ver.Id)
	if err != nil {
		fatal("submit for review", err)
	}
	log.Printf("submission %s queued — Apple will notify via email", sub.Id)
}

func fatal(label string, err error) {
	var apiErr *AppStoreConnect.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("%s: Apple rejected the request: HTTP %d, %d errors", label, apiErr.StatusCode, len(apiErr.Errors))
	}
	log.Fatalf("%s: %v", label, err)
}

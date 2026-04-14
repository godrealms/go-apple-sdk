// Example: list builds for an app, create a beta group, and
// distribute the newest build to that group.
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
	appID := "" // App Store Connect resource id of the app

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. Find the newest build for the app. Apple lets us filter by
	//    the parent app relationship and sort by uploadedDate desc.
	buildsQ := AppStoreConnect.NewQuery().
		Filter("app", appID).
		Sort("-uploadedDate").
		Limit(1)
	buildsResp, err := svc.Builds().List(ctx, buildsQ)
	if err != nil {
		fatal("list builds", err)
	}
	if len(buildsResp.Data) == 0 {
		log.Fatalf("no builds uploaded for app %s", appID)
	}
	newest := buildsResp.Data[0]
	log.Printf("newest build: %s (version %s)", newest.Id, newest.Attributes.Version)

	// 2. Create a beta group under the app.
	feedbackOn := true
	group, err := svc.BetaGroups().Create(ctx, AppStoreConnect.CreateBetaGroupRequest{
		AppID:           appID,
		Name:            "Release Candidates",
		FeedbackEnabled: &feedbackOn,
	})
	if err != nil {
		fatal("create beta group", err)
	}
	log.Printf("created beta group %s", group.Id)

	// 3. Attach the newest build to the group, distributing it.
	if err := svc.BetaGroups().AddBuilds(ctx, group.Id, []string{newest.Id}); err != nil {
		fatal("add builds", err)
	}
	log.Printf("build %s attached to group %s", newest.Id, group.Id)
}

// fatal prints API errors with a code count, local errors as-is,
// then exits. Shared across Phase 3 examples to keep the happy path
// unchanged between files.
func fatal(label string, err error) {
	var apiErr *AppStoreConnect.APIError
	if errors.As(err, &apiErr) {
		log.Fatalf("%s: Apple rejected the request: HTTP %d, %d errors", label, apiErr.StatusCode, len(apiErr.Errors))
	}
	log.Fatalf("%s: %v", label, err)
}

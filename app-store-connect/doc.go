// Package AppStoreConnect provides a typed Go client for the
// App Store Connect API (https://developer.apple.com/documentation/appstoreconnectapi).
//
// The package implements the JSON:API protocol (https://jsonapi.org)
// that Apple uses, exposing strongly typed resources, a fluent query
// builder, automatic pagination, and structured error handling.
//
// This package is normally reached through the root SDK:
//
//	c := Apple.NewClient(false, kid, iss, bid, privateKey)
//	svc := c.AppStoreConnect()
//	apps, _, err := svc.Apps().List(ctx, AppStoreConnect.NewQuery().Limit(200))
//
// It can also be used standalone by providing an [Authorizer]:
//
//	svc := AppStoreConnect.New(AppStoreConnect.Config{
//	    BaseURL:    "https://api.appstoreconnect.apple.com",
//	    Authorizer: myAuth,
//	})
package AppStoreConnect

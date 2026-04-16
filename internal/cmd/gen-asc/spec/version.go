// Package spec holds the vendored Apple App Store Connect OpenAPI
// specification and the constants that lock it in place.
//
// Whenever the spec is refreshed via scripts/update-spec.sh, these
// constants MUST be updated in the same commit so a `go test` run
// catches drift.
package spec

import (
	_ "embed"
)

//go:embed app_store_connect_api_openapi.json
var File []byte

// These constants describe the currently-vendored spec.
const (
	// SpecVersion is the "info.version" field from the OpenAPI root.
	SpecVersion = "4.3"

	// SpecSHA256 is shasum -a 256 of File. Verified by version_test.go.
	SpecSHA256 = "83f485960fc7198542461050e192cd21716274c5f3828c683fba35333482a02f"

	// SpecSource is the upstream URL the file was fetched from.
	SpecSource = "https://developer.apple.com/sample-code/app-store-connect/app-store-connect-openapi-specification.zip"

	// SpecDate is the download date in YYYY-MM-DD (UTC).
	SpecDate = "2026-04-16"
)

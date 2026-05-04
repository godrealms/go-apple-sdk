// Package AppStoreServer is the Go SDK client for Apple's App Store
// Server API.
//
// The package handles server-to-server queries Apple exposes at
// api.storekit.itunes.apple.com (production) and
// api.storekit-sandbox.itunes.apple.com (sandbox): transaction
// lookup, transaction history, refund history, subscription status,
// consumption reporting, mass renewal extension, and triggering
// test notifications. Each operation is a top-level function that
// takes a *Apple.Client and a context.Context.
//
// Example:
//
//	client := Apple.NewClient(true, kid, iss, bid, privateKey)
//	info, err := AppStoreServer.GetTransactionInfo(ctx, client, txID)
//	if err != nil {
//	    return err
//	}
//	tx, err := info.SignedTransactionInfo.Decrypt()
//
// Decrypt() and friends now perform full RFC 5280 chain validation
// against the embedded Apple Root CA G3 — see the jws/ package for
// details and for the *jws.Verifier API used to override the trust
// anchors in tests.
package AppStoreServer

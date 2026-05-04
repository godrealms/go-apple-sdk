// Package AppStoreNotifications decodes App Store Server
// Notifications V2 webhooks.
//
// Apple POSTs a JSON envelope `{"signedPayload": "<JWS>"}` to the
// developer's webhook URL. SignedPayload.DecodedPayload (or the
// convenience top-level Notifications function) verifies the JWS
// chain back to Apple Root CA G3, checks the leaf cert OID, and
// returns the decoded ResponseBodyV2DecodedPayload.
//
// Example:
//
//	payload, err := AppStoreNotifications.Notifications(rawSignedPayload)
//	if err != nil {
//	    var ve *jws.VerificationError
//	    if errors.As(err, &ve) { ... }
//	    return err
//	}
//	switch payload.NotificationType {
//	case "DID_RENEW":
//	    ...
//	}
//
// All chain validation lives in the jws/ package; use
// SignedPayload.DecodedPayloadWith to supply a custom *jws.Verifier
// (for tests with self-signed certs or future Apple root rotation).
package AppStoreNotifications

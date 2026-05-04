// Package jws verifies the JSON Web Signature payloads Apple sends
// for App Store Server Notifications V2, signed transactions, and
// signed renewal info.
//
// All verification goes through a *Verifier. Use DefaultVerifier for
// the standard configuration (Apple Root CA G3 embedded into the
// SDK, mandatory Apple receipt-signing OID check), or build a
// custom one with NewVerifier + options for testing or future
// root-cert rotation:
//
//	v := jws.NewVerifier(
//	    jws.WithRootCAs(myPool),
//	    jws.WithRequiredOIDs(myOID),
//	)
//	payload, err := jws.VerifyAndDecode[Transaction](v, raw)
//
// All failures return *VerificationError; switch on
// VerificationError.Reason to distinguish chain failure from OID
// mismatch from signature mismatch from malformed input.
//
// The package targets Apple's documented JWS profile (ES256 only,
// x5c chain present, leaf carries Apple OID). It is intentionally
// not a general-purpose JWS library.
package jws

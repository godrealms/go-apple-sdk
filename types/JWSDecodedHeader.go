package types

import "github.com/godrealms/go-apple-sdk/jws"

// X5c is the JWS x5c header field, aliased from the jws package.
//
// The previous types.X5c.GetPublicKey() method has been removed.
// It returned only the leaf cert without validating the chain,
// which made every consumer trivially impersonable. Migrate to
// jws.Verifier (or the new Decrypt / DecodedPayload methods on
// JWSTransaction / JWSRenewalInfo / SignedPayload, which all use
// jws.DefaultVerifier internally).
type X5c = jws.X5c

// JWSDecodedHeader is the JWS protected header, aliased from the
// jws package. The header's Alg field is a plain string (jws.Alg
// is `type Alg = string`); the standalone types.Alg type with a
// String() method lives in types/alg.go and is independent.
type JWSDecodedHeader = jws.Header

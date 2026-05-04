package jws

import "fmt"

// ReasonCode is the kind of failure VerifyAndDecode encountered.
// It is exposed so callers can branch on the failure category for
// alerting and metrics without parsing error strings.
type ReasonCode int

const (
	// ReasonStructure means the JWS was malformed: wrong number of
	// segments, bad base64, undecodable header or payload, or the
	// header carried an algorithm other than ES256.
	ReasonStructure ReasonCode = iota + 1
	// ReasonChain means the leaf certificate did not chain back to
	// a trusted root via the intermediates supplied in x5c.
	ReasonChain
	// ReasonOID means the leaf certificate did not carry any of
	// the OIDs the Verifier required.
	ReasonOID
	// ReasonExpired means at least one certificate in the chain was
	// outside its validity window at verification time.
	ReasonExpired
	// ReasonSignature means the leaf cert public key did not verify
	// the JWS signature, or the signature bytes were malformed
	// (wrong length, non-ECDSA leaf key).
	ReasonSignature
)

// String returns the lowercase reason name used in error messages.
func (r ReasonCode) String() string {
	switch r {
	case ReasonStructure:
		return "structure"
	case ReasonChain:
		return "chain"
	case ReasonOID:
		return "oid"
	case ReasonExpired:
		return "expired"
	case ReasonSignature:
		return "signature"
	default:
		return "unknown"
	}
}

// VerificationError is the only error type returned from this
// package. Inspect VerificationError.Reason to branch on failure
// class; use errors.Is/As against Cause for finer details (the
// wrapped error is typically an *x509 error or a JSON decode
// error).
type VerificationError struct {
	Reason ReasonCode
	Cause  error // optional; nil when the failure has no underlying error
}

// Error implements error. Format: "jws: <reason>[: <cause>]".
func (e *VerificationError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("jws: %s", e.Reason)
	}
	return fmt.Sprintf("jws: %s: %v", e.Reason, e.Cause)
}

// Unwrap returns the wrapped cause so errors.Is and errors.As walk
// through to it.
func (e *VerificationError) Unwrap() error { return e.Cause }

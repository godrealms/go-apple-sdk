package jws

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// VerifyAndDecode is the single verification entry point. It
// splits the JWS, validates the certificate chain, checks the leaf
// OID, verifies the ES256 signature, and only then decodes the
// payload as *T. Any failure returns *VerificationError; see
// ReasonCode for the categories.
//
// The signature is verified BEFORE the payload is JSON-decoded so
// we never run a JSON parser on untrusted bytes.
func VerifyAndDecode[T any](v *Verifier, raw string) (*T, error) {
	// 1. Split into 3 segments.
	segments := strings.SplitN(raw, ".", 3)
	if len(segments) != 3 {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("expected 3 JWS segments, got %d", len(segments)),
		}
	}
	headerB64, payloadB64, sigB64 := segments[0], segments[1], segments[2]

	// 2. Decode header; enforce alg=ES256.
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("header base64: %w", err),
		}
	}
	var header Header
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("header json: %w", err),
		}
	}
	if header.Alg != "ES256" {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("unsupported alg %q (only ES256)", header.Alg),
		}
	}

	// 3. Parse x5c chain.
	chain, err := header.X5c.Parse() // already returns *VerificationError
	if err != nil {
		return nil, err
	}

	// 4. Chain validation.
	if err := verifyChain(chain, v.roots, v.clock()); err != nil {
		return nil, err
	}

	// 5. OID check.
	if !matchOID(chain[0].Extensions, v.requiredOIDs) {
		return nil, &VerificationError{
			Reason: ReasonOID,
			Cause:  errors.New("leaf cert carries none of the required OIDs"),
		}
	}

	// 6. Signature verification — BEFORE payload JSON decode so we
	// never run a JSON parser on untrusted bytes.
	sig, err := base64.RawURLEncoding.DecodeString(sigB64)
	if err != nil {
		return nil, &VerificationError{
			Reason: ReasonSignature,
			Cause:  fmt.Errorf("signature base64: %w", err),
		}
	}
	if err := verifySignature(chain[0], headerB64, payloadB64, sig); err != nil {
		return nil, err
	}

	// 7. Decode payload (only after signature passes).
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("payload base64: %w", err),
		}
	}
	out := new(T)
	if err := json.Unmarshal(payloadBytes, out); err != nil {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("payload json: %w", err),
		}
	}
	return out, nil
}

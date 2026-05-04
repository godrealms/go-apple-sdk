package jws

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// X5c is the JWS x5c header field: an array of base64-std-encoded
// DER certificates. By RFC 7515 convention, X5c[0] is the leaf
// (signing) cert and the rest are intermediates in chain order.
type X5c []string

// Parse decodes each entry into a *x509.Certificate. Returns a
// *VerificationError with ReasonStructure on any failure, including
// an empty chain.
func (x X5c) Parse() ([]*x509.Certificate, error) {
	if len(x) == 0 {
		return nil, &VerificationError{
			Reason: ReasonStructure,
			Cause:  fmt.Errorf("x5c: empty chain"),
		}
	}
	out := make([]*x509.Certificate, 0, len(x))
	for i, b64 := range x {
		der, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, &VerificationError{
				Reason: ReasonStructure,
				Cause:  fmt.Errorf("x5c[%d]: base64: %w", i, err),
			}
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, &VerificationError{
				Reason: ReasonStructure,
				Cause:  fmt.Errorf("x5c[%d]: parse: %w", i, err),
			}
		}
		out = append(out, cert)
	}
	return out, nil
}

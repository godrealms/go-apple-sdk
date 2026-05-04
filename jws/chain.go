package jws

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"math/big"
	"time"
)

// verifyChain runs RFC 5280 path validation on a JWS x5c chain.
//
// chain[0] must be the leaf; chain[1:] are intermediates (in any
// order — the pool deduplicates). roots is the trust anchor pool.
// now is the verification clock; passing time.Now() is normal,
// passing a fixed instant is for tests.
//
// Returns nil on success, or *VerificationError with one of:
//
//	ReasonExpired — when the chain is rejected for time-window
//	                reasons (NotBefore in the future, NotAfter in
//	                the past)
//	ReasonChain   — for any other path-validation failure (unknown
//	                authority, name mismatch, key usage, etc.)
func verifyChain(chain []*x509.Certificate, roots *x509.CertPool, now time.Time) error {
	if len(chain) == 0 {
		return &VerificationError{
			Reason: ReasonStructure,
			Cause:  errors.New("chain: empty"),
		}
	}
	intermediates := x509.NewCertPool()
	for _, c := range chain[1:] {
		intermediates.AddCert(c)
	}
	_, err := chain[0].Verify(x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   now,
	})
	if err == nil {
		return nil
	}
	// Map x509.CertificateInvalidError{Expired} to ReasonExpired so
	// callers can branch on it. Note: x509.Expired covers BOTH
	// "NotAfter in the past" and "NotBefore in the future" (the
	// stdlib does not split them). All other x509 errors map to
	// ReasonChain.
	var invErr x509.CertificateInvalidError
	if errors.As(err, &invErr) {
		if invErr.Reason == x509.Expired {
			return &VerificationError{Reason: ReasonExpired, Cause: err}
		}
	}
	return &VerificationError{Reason: ReasonChain, Cause: err}
}

// verifySignature verifies a JWS ES256 signature using the leaf
// cert's public key.
//
// sig must be exactly 64 bytes (IEEE P1363 raw r ‖ s, each 32
// bytes). The leaf's PublicKey must be *ecdsa.PublicKey on a
// P-256 curve. Any other shape fails with ReasonSignature.
func verifySignature(leaf *x509.Certificate, headerB64, payloadB64 string, sig []byte) error {
	if len(sig) != 64 {
		return &VerificationError{
			Reason: ReasonSignature,
			Cause:  errors.New("signature: expected 64 bytes (ES256 IEEE P1363)"),
		}
	}
	pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return &VerificationError{
			Reason: ReasonSignature,
			Cause:  errors.New("signature: leaf public key is not ECDSA"),
		}
	}
	hash := sha256.Sum256([]byte(headerB64 + "." + payloadB64))
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	if !ecdsa.Verify(pub, hash[:], r, s) {
		return &VerificationError{
			Reason: ReasonSignature,
			Cause:  errors.New("signature: ECDSA verify failed"),
		}
	}
	return nil
}

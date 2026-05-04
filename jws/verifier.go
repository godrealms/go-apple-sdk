package jws

import (
	"crypto/x509"
	"encoding/asn1"
	"time"
)

// Verifier holds the trust anchors, required OIDs, and clock used
// by VerifyAndDecode. Construct with NewVerifier or grab the
// process-wide DefaultVerifier.
//
// Once constructed, a Verifier is safe for concurrent use by
// multiple goroutines. Its fields are not modified after
// construction.
type Verifier struct {
	roots        *x509.CertPool
	requiredOIDs []asn1.ObjectIdentifier
	clock        func() time.Time
}

// Option mutates a Verifier during NewVerifier.
type Option func(*Verifier)

// NewVerifier builds a Verifier. With no options it uses package
// defaults: an empty roots pool (callers MUST supply roots via
// WithRootCAs or use DefaultVerifier), DefaultRequiredOIDs as the
// required OID list, and time.Now as the clock.
//
// For production code, prefer DefaultVerifier(), which embeds
// Apple Root CA G3 automatically.
func NewVerifier(opts ...Option) *Verifier {
	v := &Verifier{
		roots:        x509.NewCertPool(),
		requiredOIDs: append([]asn1.ObjectIdentifier(nil), DefaultRequiredOIDs...),
		clock:        time.Now,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// WithRootCAs replaces the trust anchor pool. The Verifier takes
// the pool as-is; do not mutate it after construction.
func WithRootCAs(pool *x509.CertPool) Option {
	return func(v *Verifier) { v.roots = pool }
}

// WithRequiredOIDs replaces the required OID list. A leaf cert
// passes if it carries ANY of the listed OIDs (logical OR).
func WithRequiredOIDs(oids ...asn1.ObjectIdentifier) Option {
	return func(v *Verifier) {
		v.requiredOIDs = append([]asn1.ObjectIdentifier(nil), oids...)
	}
}

// WithClock replaces the verification clock. Tests use this to
// exercise expired / not-yet-valid certificate paths
// deterministically; production code should not call it.
func WithClock(now func() time.Time) Option {
	return func(v *Verifier) { v.clock = now }
}

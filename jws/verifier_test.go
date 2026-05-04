package jws

import (
	"crypto/x509"
	"encoding/asn1"
	"testing"
	"time"
)

func TestNewVerifier_Defaults(t *testing.T) {
	v := NewVerifier()
	if v.roots == nil {
		t.Fatalf("expected non-nil default roots pool (will be empty until WithRootCAs)")
	}
	if len(v.requiredOIDs) == 0 {
		t.Fatalf("expected default required OIDs to be set")
	}
	if v.clock == nil {
		t.Fatalf("expected default clock function")
	}
	// Default clock should approximate time.Now.
	got := v.clock()
	if time.Since(got) > time.Second {
		t.Fatalf("default clock returned stale time: %v", got)
	}
}

func TestNewVerifier_WithRootCAs(t *testing.T) {
	pool := x509.NewCertPool()
	v := NewVerifier(WithRootCAs(pool))
	if v.roots != pool {
		t.Fatalf("WithRootCAs did not set the pool")
	}
}

func TestNewVerifier_WithRequiredOIDs(t *testing.T) {
	myOID := asn1.ObjectIdentifier{1, 2, 3}
	v := NewVerifier(WithRequiredOIDs(myOID))
	if len(v.requiredOIDs) != 1 || !v.requiredOIDs[0].Equal(myOID) {
		t.Fatalf("WithRequiredOIDs did not replace OID list, got %v", v.requiredOIDs)
	}
}

func TestNewVerifier_WithClock(t *testing.T) {
	fixed := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	v := NewVerifier(WithClock(func() time.Time { return fixed }))
	if !v.clock().Equal(fixed) {
		t.Fatalf("WithClock did not install custom clock")
	}
}

func TestNewVerifier_OptionsCompose(t *testing.T) {
	pool := x509.NewCertPool()
	myOID := asn1.ObjectIdentifier{4, 5, 6}
	fixed := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	v := NewVerifier(
		WithRootCAs(pool),
		WithRequiredOIDs(myOID),
		WithClock(func() time.Time { return fixed }),
	)
	if v.roots != pool || !v.requiredOIDs[0].Equal(myOID) || !v.clock().Equal(fixed) {
		t.Fatalf("options did not all apply")
	}
}

package jws

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"
)

func TestMatchOID_Found(t *testing.T) {
	exts := []pkix.Extension{{Id: OIDAppleReceiptSigning}}
	if !matchOID(exts, []asn1.ObjectIdentifier{OIDAppleReceiptSigning}) {
		t.Fatalf("expected match")
	}
}

func TestMatchOID_NotFound(t *testing.T) {
	exts := []pkix.Extension{{Id: asn1.ObjectIdentifier{1, 2, 3}}}
	if matchOID(exts, []asn1.ObjectIdentifier{OIDAppleReceiptSigning}) {
		t.Fatalf("expected no match")
	}
}

func TestMatchOID_AnyOfRequired(t *testing.T) {
	exts := []pkix.Extension{{Id: OIDAppleNotificationSigning}}
	// Required list contains both — any one is enough.
	if !matchOID(exts, []asn1.ObjectIdentifier{
		OIDAppleReceiptSigning, OIDAppleNotificationSigning,
	}) {
		t.Fatalf("expected match (notification OID present)")
	}
}

func TestMatchOID_NoExtensions(t *testing.T) {
	if matchOID(nil, []asn1.ObjectIdentifier{OIDAppleReceiptSigning}) {
		t.Fatalf("expected no match against empty extensions")
	}
}

func TestMatchOID_NoRequired(t *testing.T) {
	exts := []pkix.Extension{{Id: OIDAppleReceiptSigning}}
	if matchOID(exts, nil) {
		t.Fatalf("expected no match against empty required list")
	}
}

func TestDefaultRequiredOIDs_ContainsReceiptSigning(t *testing.T) {
	// Conservative default ships with ONLY the receipt-signing OID.
	// If this test fails, someone added another OID without
	// confirming it against a real Apple capture.
	if len(DefaultRequiredOIDs) != 1 {
		t.Fatalf("DefaultRequiredOIDs should contain exactly 1 OID (conservative); got %d", len(DefaultRequiredOIDs))
	}
	if !DefaultRequiredOIDs[0].Equal(OIDAppleReceiptSigning) {
		t.Fatalf("DefaultRequiredOIDs[0] should be OIDAppleReceiptSigning")
	}
}

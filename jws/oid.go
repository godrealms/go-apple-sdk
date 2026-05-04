package jws

import (
	"crypto/x509/pkix"
	"encoding/asn1"
)

// OIDAppleReceiptSigning is the X.509 OID Apple stamps on every
// receipt-signing leaf certificate. Documented since iOS 7 and
// confirmed present on App Store Server Notifications V2 leaves.
var OIDAppleReceiptSigning = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}

// OIDAppleNotificationSigning is the suspected (but not yet
// independently verified) OID for App Store Server Notifications
// V2 signing. Kept as a defined constant so callers can opt in via
// WithRequiredOIDs once they have confirmed it on their own
// sandbox capture, but DELIBERATELY omitted from
// DefaultRequiredOIDs below — including an unverified OID in the
// default list would risk rejecting every legitimate notification.
var OIDAppleNotificationSigning = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 29}

// DefaultRequiredOIDs lists the OIDs DefaultVerifier requires the
// leaf cert to carry. A leaf passes if its Extensions list contains
// ANY of these (logical OR), not all of them.
//
// Conservatively we ship with only OIDAppleReceiptSigning, which
// Apple's own documentation confirms is present on every leaf they
// use to sign receipts, transactions, and notifications. To
// require additional OIDs (e.g. the suspected V2 notification OID
// 1.2.840.113635.100.6.29 once you've verified it against a real
// sandbox notification), construct a Verifier with
// WithRequiredOIDs(OIDAppleReceiptSigning, OIDAppleNotificationSigning, ...).
var DefaultRequiredOIDs = []asn1.ObjectIdentifier{
	OIDAppleReceiptSigning,
}

// matchOID returns true if any extension's Id appears in required.
// matchOID is a free function (not a method on *x509.Certificate)
// so unit tests can pass synthetic extension lists.
func matchOID(extensions []pkix.Extension, required []asn1.ObjectIdentifier) bool {
	for _, ext := range extensions {
		for _, req := range required {
			if ext.Id.Equal(req) {
				return true
			}
		}
	}
	return false
}

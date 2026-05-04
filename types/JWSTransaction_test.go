package types

import (
	"errors"
	"testing"

	"github.com/godrealms/go-apple-sdk/jws"
	"github.com/godrealms/go-apple-sdk/internal/testchain"
)

func TestJWSTransaction_Decrypt_DefaultVerifierRejectsTestChain(t *testing.T) {
	// The default Verifier requires the real Apple Root CA, so a
	// testchain payload must fail with ReasonChain. That's the
	// signal we want: it proves Decrypt is now hitting the
	// verifier path (rather than blindly accepting any cert as
	// before the fix).
	tc := testchain.New(t)
	raw := tc.SignJWS(t, JWSTransactionDecodedPayload{TransactionId: "abc"})
	_, err := JWSTransaction(raw).Decrypt()
	var ve *jws.VerificationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *jws.VerificationError, got %T: %v", err, err)
	}
	if ve.Reason != jws.ReasonChain {
		t.Fatalf("expected ReasonChain, got %s", ve.Reason)
	}
}

func TestJWSTransaction_DecryptWith_AcceptsTestVerifier(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, JWSTransactionDecodedPayload{TransactionId: "tx-123"})
	v := jws.NewVerifier(
		jws.WithRootCAs(tc.RootPool),
		jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning),
	)
	out, err := JWSTransaction(raw).DecryptWith(v)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if string(out.TransactionId) != "tx-123" {
		t.Fatalf("transaction id = %q, want tx-123", out.TransactionId)
	}
}

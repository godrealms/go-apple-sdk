package types

import (
	"errors"
	"testing"

	"github.com/godrealms/go-apple-sdk/jws"
	"github.com/godrealms/go-apple-sdk/internal/testchain"
)

func TestJWSRenewalInfo_Decrypt_DefaultVerifierRejectsTestChain(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, JWSRenewalInfoDecodedPayload{ProductId: "p"})
	_, err := JWSRenewalInfo(raw).Decrypt()
	var ve *jws.VerificationError
	if !errors.As(err, &ve) || ve.Reason != jws.ReasonChain {
		t.Fatalf("expected ReasonChain, got %v", err)
	}
}

func TestJWSRenewalInfo_DecryptWith_AcceptsTestVerifier(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, JWSRenewalInfoDecodedPayload{ProductId: "p-9"})
	v := jws.NewVerifier(
		jws.WithRootCAs(tc.RootPool),
		jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning),
	)
	out, err := JWSRenewalInfo(raw).DecryptWith(v)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if string(out.ProductId) != "p-9" {
		t.Fatalf("product id = %q, want p-9", out.ProductId)
	}
}

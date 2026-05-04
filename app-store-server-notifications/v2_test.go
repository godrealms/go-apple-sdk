package AppStoreNotifications

import (
	"errors"
	"testing"

	"github.com/godrealms/go-apple-sdk/jws"
	"github.com/godrealms/go-apple-sdk/internal/testchain"
)

func TestSignedPayload_DecodedPayload_DefaultVerifierRejectsTestChain(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, ResponseBodyV2DecodedPayload{NotificationUUID: "u-1"})
	_, err := SignedPayload(raw).DecodedPayload()
	var ve *jws.VerificationError
	if !errors.As(err, &ve) || ve.Reason != jws.ReasonChain {
		t.Fatalf("expected ReasonChain, got %v", err)
	}
}

func TestSignedPayload_DecodedPayloadWith_AcceptsTestVerifier(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, ResponseBodyV2DecodedPayload{NotificationUUID: "u-42"})
	v := jws.NewVerifier(
		jws.WithRootCAs(tc.RootPool),
		jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning),
	)
	out, err := SignedPayload(raw).DecodedPayloadWith(v)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if string(out.NotificationUUID) != "u-42" {
		t.Fatalf("notification uuid = %q, want u-42", out.NotificationUUID)
	}
}

func TestNotifications_DelegatesToDefaultVerifier(t *testing.T) {
	// Notifications() is a thin wrapper that calls
	// SignedPayload.DecodedPayload internally. Make sure the
	// helper actually uses the verifier path (rather than
	// regressing back to the unsafe legacy logic).
	tc := testchain.New(t)
	raw := tc.SignJWS(t, ResponseBodyV2DecodedPayload{NotificationUUID: "u-x"})
	_, err := Notifications(raw)
	var ve *jws.VerificationError
	if !errors.As(err, &ve) || ve.Reason != jws.ReasonChain {
		t.Fatalf("expected ReasonChain via Notifications, got %v", err)
	}
}

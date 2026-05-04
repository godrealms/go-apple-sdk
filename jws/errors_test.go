package jws

import (
	"errors"
	"fmt"
	"testing"
)

func TestVerificationError_Error(t *testing.T) {
	cases := []struct {
		name string
		err  *VerificationError
		want string
	}{
		{"with cause",
			&VerificationError{Reason: ReasonChain, Cause: fmt.Errorf("boom")},
			"jws: chain: boom"},
		{"without cause",
			&VerificationError{Reason: ReasonOID},
			"jws: oid"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.err.Error(); got != c.want {
				t.Fatalf("Error() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestVerificationError_Unwrap(t *testing.T) {
	sentinel := fmt.Errorf("inner")
	ve := &VerificationError{Reason: ReasonSignature, Cause: sentinel}
	if !errors.Is(ve, sentinel) {
		t.Fatalf("errors.Is should match wrapped cause")
	}
}

func TestReasonCode_String(t *testing.T) {
	cases := map[ReasonCode]string{
		ReasonStructure: "structure",
		ReasonChain:     "chain",
		ReasonOID:       "oid",
		ReasonExpired:   "expired",
		ReasonSignature: "signature",
		ReasonCode(99):  "unknown",
	}
	for code, want := range cases {
		if got := code.String(); got != want {
			t.Fatalf("ReasonCode(%d).String() = %q, want %q", code, got, want)
		}
	}
}

// assertReason is the shared helper for verifying that a returned
// error is a *VerificationError with the expected Reason. Defined
// here, used across the package's test files.
func assertReason(t *testing.T, err error, want ReasonCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with reason %s, got nil", want)
	}
	var ve *VerificationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *VerificationError, got %T: %v", err, err)
	}
	if ve.Reason != want {
		t.Fatalf("expected reason %s, got %s (cause=%v)", want, ve.Reason, ve.Cause)
	}
}

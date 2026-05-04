package jws

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-apple-sdk/internal/testchain"
)

// verifyChain tests

func TestVerifyChain_Success(t *testing.T) {
	tc := testchain.New(t)
	err := verifyChain(
		[]*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
		tc.RootPool,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestVerifyChain_WrongRoot(t *testing.T) {
	tc := testchain.New(t)
	otherRoots := x509.NewCertPool() // empty
	err := verifyChain(
		[]*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
		otherRoots,
		time.Now(),
	)
	assertReason(t, err, ReasonChain)
}

func TestVerifyChain_LeafExpired(t *testing.T) {
	tc := testchain.New(t, testchain.WithLeafNotAfter(time.Now().Add(-time.Hour)))
	err := verifyChain(
		[]*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
		tc.RootPool,
		time.Now(),
	)
	assertReason(t, err, ReasonExpired)
}

func TestVerifyChain_LeafNotYetValid(t *testing.T) {
	tc := testchain.New(t, testchain.WithLeafNotBefore(time.Now().Add(time.Hour)))
	err := verifyChain(
		[]*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
		tc.RootPool,
		time.Now(),
	)
	assertReason(t, err, ReasonExpired)
}

func TestVerifyChain_EmptyChain(t *testing.T) {
	err := verifyChain(nil, x509.NewCertPool(), time.Now())
	assertReason(t, err, ReasonStructure)
}

// verifySignature tests

func TestVerifySignature_Success(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, map[string]string{"hello": "world"})
	bits := strings.SplitN(raw, ".", 3)
	sig, err := base64.RawURLEncoding.DecodeString(bits[2])
	if err != nil {
		t.Fatalf("decode sig: %v", err)
	}
	if err := verifySignature(tc.Leaf, bits[0], bits[1], sig); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestVerifySignature_TamperedSigningInput(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, map[string]string{"hello": "world"})
	bits := strings.SplitN(raw, ".", 3)
	sig, _ := base64.RawURLEncoding.DecodeString(bits[2])
	// Substitute a different payloadB64 — same length, different content.
	otherPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"hi":"x"}`))
	err := verifySignature(tc.Leaf, bits[0], otherPayload, sig)
	assertReason(t, err, ReasonSignature)
}

func TestVerifySignature_TruncatedSignature(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, map[string]string{"hello": "world"})
	bits := strings.SplitN(raw, ".", 3)
	sig, _ := base64.RawURLEncoding.DecodeString(bits[2])
	err := verifySignature(tc.Leaf, bits[0], bits[1], sig[:60])
	assertReason(t, err, ReasonSignature)
}

func TestVerifySignature_NonECDSAKey(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa keygen: %v", err)
	}
	cert := &x509.Certificate{PublicKey: &rsaKey.PublicKey}
	err = verifySignature(cert, "header", "payload", make([]byte, 64))
	assertReason(t, err, ReasonSignature)
}

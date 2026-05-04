package testchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"
)

func TestNew_BuildsValidChain(t *testing.T) {
	tc := New(t)
	if tc.Root == nil || tc.Intermediate == nil || tc.Leaf == nil {
		t.Fatalf("expected all three certs populated")
	}
	if err := tc.Intermediate.CheckSignatureFrom(tc.Root); err != nil {
		t.Fatalf("intermediate not signed by root: %v", err)
	}
	if err := tc.Leaf.CheckSignatureFrom(tc.Intermediate); err != nil {
		t.Fatalf("leaf not signed by intermediate: %v", err)
	}
	if tc.RootPool == nil {
		t.Fatalf("expected RootPool populated")
	}
}

func TestSignJWS_RoundTripVerifies(t *testing.T) {
	tc := New(t)
	type payload struct {
		Foo string `json:"foo"`
	}
	raw := tc.SignJWS(t, payload{Foo: "bar"})
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWS segments, got %d", len(parts))
	}
	headerJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
	var hdr struct {
		Alg string   `json:"alg"`
		X5c []string `json:"x5c"`
	}
	if err := json.Unmarshal(headerJSON, &hdr); err != nil {
		t.Fatalf("decode header: %v", err)
	}
	if hdr.Alg != "ES256" {
		t.Fatalf("expected alg ES256, got %q", hdr.Alg)
	}
	sig, _ := base64.RawURLEncoding.DecodeString(parts[2])
	if len(sig) != 64 {
		t.Fatalf("expected 64-byte raw signature, got %d", len(sig))
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	pub := tc.Leaf.PublicKey.(*ecdsa.PublicKey)
	if !ecdsa.Verify(pub, hash[:], r, s) {
		t.Fatalf("ecdsa verify failed on testchain output")
	}
}

func TestNew_WithLeafOIDs_AppliesExtensions(t *testing.T) {
	myOID := asn1.ObjectIdentifier{1, 2, 3, 4, 5}
	tc := New(t, WithLeafOIDs(myOID))
	var found bool
	for _, ext := range tc.Leaf.Extensions {
		if ext.Id.Equal(myOID) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected custom OID in leaf extensions")
	}
}

func TestNew_WithLeafNotAfter_AppliesExpiration(t *testing.T) {
	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	tc := New(t, WithLeafNotAfter(past))
	if !tc.Leaf.NotAfter.Equal(past) {
		t.Fatalf("expected leaf NotAfter %v, got %v", past, tc.Leaf.NotAfter)
	}
}

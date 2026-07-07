package types

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func mustPKCS8PEM(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func mustECPEM(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
}

func TestParsePrivateKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("PKCS8", func(t *testing.T) {
		got, err := ParsePrivateKey(mustPKCS8PEM(t, key))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Equal(key) {
			t.Error("parsed PKCS#8 key does not match original")
		}
	})
	t.Run("EC", func(t *testing.T) {
		got, err := ParsePrivateKey(mustECPEM(t, key))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.Equal(key) {
			t.Error("parsed EC key does not match original")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if _, err := ParsePrivateKey(""); err == nil {
			t.Error("expected error for empty input")
		}
	})
	t.Run("garbage", func(t *testing.T) {
		if _, err := ParsePrivateKey("not a pem block"); err == nil {
			t.Error("expected error for non-PEM input")
		}
	})
	t.Run("unsupported_type", func(t *testing.T) {
		block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: []byte{1, 2, 3}}
		if _, err := ParsePrivateKey(string(pem.EncodeToMemory(block))); err == nil {
			t.Error("expected error for unsupported key type")
		}
	})
}

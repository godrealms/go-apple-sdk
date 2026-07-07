package Apple

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/godrealms/go-apple-sdk/v2/types"
)

func genConfigKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func pkcs8Base64Body(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(der)
}

func fullPKCS8PEM(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func fullECPEM(t *testing.T, key *ecdsa.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}))
}

// TestNewConfig_WrapsBareBase64 verifies a bare base64 PKCS#8 body is wrapped
// into a valid PEM that ParsePrivateKey accepts.
func TestNewConfig_WrapsBareBase64(t *testing.T) {
	key := genConfigKey(t)
	cfg := NewConfig("kid", "iss", "bid", pkcs8Base64Body(t, key))

	if !strings.HasPrefix(cfg.PrivateKey, "-----BEGIN PRIVATE KEY-----") {
		t.Errorf("missing PEM header: %q", cfg.PrivateKey)
	}
	if !strings.HasSuffix(strings.TrimSpace(cfg.PrivateKey), "-----END PRIVATE KEY-----") {
		t.Errorf("missing PEM footer: %q", cfg.PrivateKey)
	}
	if _, err := types.ParsePrivateKey(cfg.PrivateKey); err != nil {
		t.Errorf("wrapped key should parse: %v", err)
	}
}

// TestNewConfig_PreservesFullPEM guards the double-wrap bug: a caller who
// already supplies a complete PEM block (PKCS#8 or SEC1 "EC PRIVATE KEY")
// must get it back intact and parseable, not corrupted by extra headers.
func TestNewConfig_PreservesFullPEM(t *testing.T) {
	key := genConfigKey(t)
	for name, pemStr := range map[string]string{
		"pkcs8": fullPKCS8PEM(t, key),
		"ec":    fullECPEM(t, key),
	} {
		t.Run(name, func(t *testing.T) {
			cfg := NewConfig("kid", "iss", "bid", pemStr)
			if _, err := types.ParsePrivateKey(cfg.PrivateKey); err != nil {
				t.Errorf("full %s PEM should survive NewConfig and parse, got: %v\nkey:\n%s",
					name, err, cfg.PrivateKey)
			}
		})
	}
}

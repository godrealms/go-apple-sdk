// Package testchain builds in-memory ECDSA certificate chains and
// signs JWS payloads with them, for use by jws/ tests and by tests
// in dependent packages (e.g. types/JWSTransaction migration tests).
//
// The package lives under internal/ on purpose: it is for SDK tests
// only, never for downstream callers' production code.
package testchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

// Chain is the output of New. It carries enough material to either
// (a) sign a JWS via SignJWS, or (b) hand its RootPool to a
// Verifier under test.
type Chain struct {
	Root, Intermediate, Leaf *x509.Certificate
	LeafKey                  *ecdsa.PrivateKey
	RootPool                 *x509.CertPool
}

// Opt customises the leaf cert built by New.
type Opt func(*config)

type config struct {
	leafOIDs      []asn1.ObjectIdentifier
	leafNotBefore time.Time
	leafNotAfter  time.Time
}

// WithLeafOIDs adds the given OIDs as custom X.509 extensions on
// the leaf cert. Each is encoded with an empty ASN.1 NULL value,
// which matches how Apple writes its receipt-signing OID extension.
func WithLeafOIDs(oids ...asn1.ObjectIdentifier) Opt {
	return func(c *config) { c.leafOIDs = oids }
}

// WithLeafNotBefore overrides the leaf cert's NotBefore.
func WithLeafNotBefore(t time.Time) Opt {
	return func(c *config) { c.leafNotBefore = t }
}

// WithLeafNotAfter overrides the leaf cert's NotAfter.
func WithLeafNotAfter(t time.Time) Opt {
	return func(c *config) { c.leafNotAfter = t }
}

// appleReceiptSigningOID is the default OID stamped on the test
// leaf — matches the production Apple receipt-signing OID so the
// jws.DefaultVerifier accepts test-chain payloads without
// additional configuration. Tests that want to exercise the OID
// failure path should override with WithLeafOIDs.
var appleReceiptSigningOID = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}

// New builds a fresh root → intermediate → leaf chain. All keys
// and certs live in process memory only; nothing touches disk.
func New(t *testing.T, opts ...Opt) *Chain {
	t.Helper()
	cfg := &config{
		leafOIDs:      []asn1.ObjectIdentifier{appleReceiptSigningOID},
		leafNotBefore: time.Now().Add(-time.Hour),
		leafNotAfter:  time.Now().Add(24 * time.Hour),
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Root: self-signed P-384 (mirrors Apple Root CA G3 curve).
	rootKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	must(t, err, "generate root key")
	rootTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test root"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTpl, rootTpl, &rootKey.PublicKey, rootKey)
	must(t, err, "create root cert")
	root, err := x509.ParseCertificate(rootDER)
	must(t, err, "parse root cert")

	// Intermediate: P-256, signed by root.
	intKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	must(t, err, "generate intermediate key")
	intTpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "test intermediate"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	intDER, err := x509.CreateCertificate(rand.Reader, intTpl, root, &intKey.PublicKey, rootKey)
	must(t, err, "create intermediate cert")
	intermediate, err := x509.ParseCertificate(intDER)
	must(t, err, "parse intermediate cert")

	// Leaf: P-256, signed by intermediate.
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	must(t, err, "generate leaf key")
	var extras []pkix.Extension
	for _, oid := range cfg.leafOIDs {
		extras = append(extras, pkix.Extension{
			Id:       oid,
			Critical: false,
			// ASN.1 NULL value (0x05 0x00) — minimal valid encoding.
			Value: []byte{0x05, 0x00},
		})
	}
	leafTpl := &x509.Certificate{
		SerialNumber:    big.NewInt(3),
		Subject:         pkix.Name{CommonName: "test leaf"},
		NotBefore:       cfg.leafNotBefore,
		NotAfter:        cfg.leafNotAfter,
		KeyUsage:        x509.KeyUsageDigitalSignature,
		ExtraExtensions: extras,
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTpl, intermediate, &leafKey.PublicKey, intKey)
	must(t, err, "create leaf cert")
	leaf, err := x509.ParseCertificate(leafDER)
	must(t, err, "parse leaf cert")

	pool := x509.NewCertPool()
	pool.AddCert(root)

	return &Chain{
		Root:         root,
		Intermediate: intermediate,
		Leaf:         leaf,
		LeafKey:      leafKey,
		RootPool:     pool,
	}
}

// SignJWS builds a JWS string (header.payload.signature) using the
// leaf key. payload is JSON-marshalled; header carries alg=ES256
// and x5c=[leaf, intermediate, root] (each base64-std-encoded
// DER).
func (c *Chain) SignJWS(t *testing.T, payload any) string {
	t.Helper()
	header := struct {
		Alg string   `json:"alg"`
		X5c []string `json:"x5c"`
	}{
		Alg: "ES256",
		X5c: []string{
			base64.StdEncoding.EncodeToString(c.Leaf.Raw),
			base64.StdEncoding.EncodeToString(c.Intermediate.Raw),
			base64.StdEncoding.EncodeToString(c.Root.Raw),
		},
	}
	headerJSON, err := json.Marshal(header)
	must(t, err, "marshal header")
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	payloadJSON, err := json.Marshal(payload)
	must(t, err, "marshal payload")
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := headerB64 + "." + payloadB64
	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, c.LeafKey, hash[:])
	must(t, err, "ecdsa sign")
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return signingInput + "." + sigB64
}

func must(t *testing.T, err error, what string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

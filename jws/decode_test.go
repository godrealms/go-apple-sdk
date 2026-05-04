package jws

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/godrealms/go-apple-sdk/internal/testchain"
)

type tp struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

// newTestVerifier returns a Verifier configured to trust the
// chain's root and require the chain's stamped OID.
func newTestVerifier(tc *testchain.Chain) *Verifier {
	return NewVerifier(
		WithRootCAs(tc.RootPool),
		WithRequiredOIDs(OIDAppleReceiptSigning),
	)
}

// 1. Happy path
func TestVerifyAndDecode_Success(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, tp{Foo: "hello", Bar: 42})
	out, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if out.Foo != "hello" || out.Bar != 42 {
		t.Fatalf("decoded payload = %+v, want {hello 42}", out)
	}
}

// 2. Wrong root
func TestVerifyAndDecode_UnknownRoot(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, tp{})
	v := NewVerifier(
		WithRootCAs(x509.NewCertPool()), // empty
		WithRequiredOIDs(OIDAppleReceiptSigning),
	)
	_, err := VerifyAndDecode[tp](v, raw)
	assertReason(t, err, ReasonChain)
}

// 3. Leaf expired (NotAfter in the past)
func TestVerifyAndDecode_LeafExpired(t *testing.T) {
	tc := testchain.New(t, testchain.WithLeafNotAfter(time.Now().Add(-time.Hour)))
	raw := tc.SignJWS(t, tp{})
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
	assertReason(t, err, ReasonExpired)
}

// 4. Leaf not yet valid (NotBefore in the future)
func TestVerifyAndDecode_LeafNotYetValid(t *testing.T) {
	tc := testchain.New(t, testchain.WithLeafNotBefore(time.Now().Add(time.Hour)))
	raw := tc.SignJWS(t, tp{})
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
	assertReason(t, err, ReasonExpired)
}

// 5. Leaf missing required OID
func TestVerifyAndDecode_MissingOID(t *testing.T) {
	tc := testchain.New(t, testchain.WithLeafOIDs(asn1.ObjectIdentifier{1, 2, 3, 4}))
	raw := tc.SignJWS(t, tp{})
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
	assertReason(t, err, ReasonOID)
}

// 6. Tampered payload
func TestVerifyAndDecode_TamperedPayload(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, tp{Foo: "x"})
	bits := strings.SplitN(raw, ".", 3)
	bits[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"foo":"y"}`))
	tampered := strings.Join(bits, ".")
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), tampered)
	assertReason(t, err, ReasonSignature)
}

// 7. Truncated signature
func TestVerifyAndDecode_TruncatedSignature(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, tp{})
	bits := strings.SplitN(raw, ".", 3)
	sig, _ := base64.RawURLEncoding.DecodeString(bits[2])
	bits[2] = base64.RawURLEncoding.EncodeToString(sig[:60])
	mangled := strings.Join(bits, ".")
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), mangled)
	assertReason(t, err, ReasonSignature)
}

// 8. alg != ES256
func TestVerifyAndDecode_WrongAlg(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, tp{})
	bits := strings.SplitN(raw, ".", 3)
	hdr, _ := base64.RawURLEncoding.DecodeString(bits[0])
	var hh map[string]any
	_ = json.Unmarshal(hdr, &hh)
	hh["alg"] = "RS256"
	newHdr, _ := json.Marshal(hh)
	bits[0] = base64.RawURLEncoding.EncodeToString(newHdr)
	mangled := strings.Join(bits, ".")
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), mangled)
	assertReason(t, err, ReasonStructure)
}

// 9. Wrong segment count
func TestVerifyAndDecode_TwoSegments(t *testing.T) {
	_, err := VerifyAndDecode[tp](NewVerifier(), "abc.def")
	assertReason(t, err, ReasonStructure)
}

// 10. Empty x5c
func TestVerifyAndDecode_EmptyX5c(t *testing.T) {
	hdr, _ := json.Marshal(Header{Alg: "ES256", X5c: X5c{}})
	payload, _ := json.Marshal(tp{})
	raw := base64.RawURLEncoding.EncodeToString(hdr) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(make([]byte, 64))
	_, err := VerifyAndDecode[tp](NewVerifier(), raw)
	assertReason(t, err, ReasonStructure)
}

// 11. Bad base64 inside x5c
func TestVerifyAndDecode_BadX5cBase64(t *testing.T) {
	hdr, _ := json.Marshal(Header{Alg: "ES256", X5c: X5c{"!!!not-b64!!!"}})
	payload, _ := json.Marshal(tp{})
	raw := base64.RawURLEncoding.EncodeToString(hdr) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(make([]byte, 64))
	_, err := VerifyAndDecode[tp](NewVerifier(), raw)
	assertReason(t, err, ReasonStructure)
}

// 12. Malformed JSON header
func TestVerifyAndDecode_MalformedHeaderJSON(t *testing.T) {
	mangled := base64.RawURLEncoding.EncodeToString([]byte("not json")) + ".aa.bb"
	_, err := VerifyAndDecode[tp](NewVerifier(), mangled)
	assertReason(t, err, ReasonStructure)
}

// 13. JSON payload that won't unmarshal into target type. The
// signature verifies (testchain signed over the bytes), but
// json.Unmarshal into *tp fails because "bar" is a string here.
func TestVerifyAndDecode_PayloadTypeMismatch(t *testing.T) {
	tc := testchain.New(t)
	raw := tc.SignJWS(t, map[string]any{
		"foo": "x",
		"bar": "not a number", // tp.Bar is int → unmarshal fails
	})
	_, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
	assertReason(t, err, ReasonStructure)
}

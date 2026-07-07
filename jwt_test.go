package Apple

import (
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateAppStoreServerAuthorizationJWT_Claims verifies the App Store
// Server JWT carries the correct claims/headers and an expiry within Apple's
// limit, so a signing regression surfaces as a local test failure rather than
// an opaque 401 from Apple.
func TestGenerateAppStoreServerAuthorizationJWT_Claims(t *testing.T) {
	c := NewClient(false, "KID123", "ISS456", "com.example.app", testPrivateKeyPEM(t))
	auth, err := c.GenerateAppStoreServerAuthorizationJWT()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !strings.HasPrefix(auth, "Bearer ") {
		t.Fatalf("missing Bearer prefix: %q", auth)
	}
	tok, _, err := jwt.NewParser().ParseUnverified(strings.TrimPrefix(auth, "Bearer "), jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	claims := tok.Claims.(jwt.MapClaims)
	if claims["iss"] != "ISS456" {
		t.Errorf("iss = %v, want ISS456", claims["iss"])
	}
	if claims["aud"] != "appstoreconnect-v1" {
		t.Errorf("aud = %v, want appstoreconnect-v1", claims["aud"])
	}
	if claims["bid"] != "com.example.app" {
		t.Errorf("bid = %v, want com.example.app", claims["bid"])
	}
	if tok.Header["kid"] != "KID123" {
		t.Errorf("kid header = %v, want KID123", tok.Header["kid"])
	}
	if tok.Header["alg"] != "ES256" {
		t.Errorf("alg header = %v, want ES256", tok.Header["alg"])
	}
	iat := int64(claims["iat"].(float64))
	exp := int64(claims["exp"].(float64))
	if d := exp - iat; d <= 0 || d > 60*60 {
		t.Errorf("token lifetime %ds out of Apple's (0, 60min] window", d)
	}
}

// TestSignAppStoreConnectToken_Unscoped verifies the unscoped App Store
// Connect token: no "scope" claim and an expiry within Apple's 20-minute cap.
func TestSignAppStoreConnectToken_Unscoped(t *testing.T) {
	c := NewClient(false, "KID123", "ISS456", "com.example.app", testPrivateKeyPEM(t))
	raw, err := c.signAppStoreConnectToken()
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	tok, _, err := jwt.NewParser().ParseUnverified(raw, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	claims := tok.Claims.(jwt.MapClaims)
	if _, ok := claims["scope"]; ok {
		t.Error("unscoped token must not carry a scope claim")
	}
	iat := int64(claims["iat"].(float64))
	exp := int64(claims["exp"].(float64))
	if d := exp - iat; d <= 0 || d > 20*60 {
		t.Errorf("token lifetime %ds out of Apple's (0, 20min] window", d)
	}
}

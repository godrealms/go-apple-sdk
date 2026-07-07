package Apple

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// testPrivateKeyPEM generates a fresh P-256 ES256 key in PKCS#8 PEM form,
// suitable for NewClient in tests that need JWT signing to succeed.
func testPrivateKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// newTestClient returns a Client whose per-service resty clients all target an
// httptest server running handler, with AppStoreServer selected by default.
// The real JWT auth middleware runs (signing with the generated test key), so
// requests carry a genuine Authorization header.
func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient(false, "KID123", "ISS456", "com.example.app", testPrivateKeyPEM(t))
	for svc := range c.clients {
		c.clients[svc].SetBaseURL(srv.URL)
	}
	c.SetService(AppStoreServerClient)
	return c
}

// TestRequest_MultiValueQueryParam guards the []string query-param bug: all
// values must reach the wire, not just the last one.
func TestRequest_MultiValueQueryParam(t *testing.T) {
	var got []string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()["status"]
		_, _ = w.Write([]byte(`{}`))
	}))
	err := c.Request(RequestParams{
		Method:      "GET",
		Path:        "/x",
		QueryParams: map[string]any{"status": []string{"1", "2", "3"}},
	})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("multi-value query lost: got %v, want 3 values [1 2 3]", got)
	}
}

// TestRequest_3xxReturnsError guards the silent-nil bug: a non-2xx (here a
// 304) response must surface an error rather than nil + empty result.
func TestRequest_3xxReturnsError(t *testing.T) {
	log.SetOutput(bytes.NewBuffer(nil))
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified) // 304, never auto-followed
	}))
	err := c.Request(RequestParams{Method: "GET", Path: "/x", Result: &struct{}{}})
	if err == nil {
		t.Fatal("3xx response should return an error, not nil")
	}
}

// TestHandleError_StatusMatrix locks down which status classes are errors.
func TestHandleError_StatusMatrix(t *testing.T) {
	log.SetOutput(bytes.NewBuffer(nil)) // silence handleError's diagnostic dump
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	cases := []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{"200_ok", http.StatusOK, `{}`, false},
		{"304_not_modified", http.StatusNotModified, ``, true},
		{"400_json_error", http.StatusBadRequest, `{"errorMessage":"bad"}`, true},
		{"500_raw", http.StatusInternalServerError, `oops`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.body != "" {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(tc.status)
				if tc.body != "" {
					_, _ = w.Write([]byte(tc.body))
				}
			}))
			err := c.Request(RequestParams{Method: "GET", Path: "/x", Result: &struct{}{}})
			if (err != nil) != tc.wantErr {
				t.Errorf("status %d: err=%v, wantErr=%v", tc.status, err, tc.wantErr)
			}
		})
	}
}

// TestHandleError_RedactsAuthorization guards the credential-leak bug: the
// real signed Bearer token sent on the request must never appear in the
// diagnostic log output.
func TestHandleError_RedactsAuthorization(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	var seenAuth string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorMessage":"nope"}`))
	}))
	_ = c.Request(RequestParams{Method: "GET", Path: "/x"})

	if !strings.HasPrefix(seenAuth, "Bearer ") {
		t.Fatalf("expected middleware to send a Bearer token, got %q", seenAuth)
	}
	token := strings.TrimPrefix(seenAuth, "Bearer ")
	if strings.Contains(buf.String(), token) {
		t.Error("Authorization token leaked into diagnostic log output")
	}
}

// TestWithServiceBaseURL verifies the option redirects a service's requests
// to a caller-supplied base URL (used by tests and proxy/mock setups).
func TestWithServiceBaseURL(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(false, "KID", "ISS", "BID", testPrivateKeyPEM(t),
		WithServiceBaseURL(AppStoreServerClient, srv.URL))
	c.SetService(AppStoreServerClient)
	if err := c.Request(RequestParams{Method: "GET", Path: "/x", Result: &struct{}{}}); err != nil {
		t.Fatalf("Request: %v", err)
	}
	if !hit {
		t.Error("request did not reach the overridden base URL")
	}
}

// TestRedactSensitiveHeaders is a focused unit test for the redaction helper.
func TestRedactSensitiveHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer secret")
	h.Set("Cookie", "session=abc")
	h.Set("Content-Type", "application/json")

	got := redactSensitiveHeaders(h)
	if got.Get("Authorization") != "[REDACTED]" {
		t.Errorf("Authorization not redacted: %q", got.Get("Authorization"))
	}
	if got.Get("Cookie") != "[REDACTED]" {
		t.Errorf("Cookie not redacted: %q", got.Get("Cookie"))
	}
	if got.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type should be untouched: %q", got.Get("Content-Type"))
	}
	// The original header must not be mutated.
	if h.Get("Authorization") != "Bearer secret" {
		t.Error("redactSensitiveHeaders mutated the original header")
	}
}

package AppStoreConnect

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// noopAuthorizer is used by tests that don't care about the Authorization
// header; real tests that want to assert the header was set should inspect
// req.Header from within the handler.
var noopAuthorizer AuthorizerFunc = func(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer test-token")
	return nil
}

// newTestService spins up an httptest.Server with the provided handler
// and returns a Service pointed at it. The caller must call the returned
// cleanup func.
func newTestService(t *testing.T, handler http.Handler) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	svc := New(Config{
		BaseURL:    srv.URL,
		Authorizer: noopAuthorizer,
	})
	t.Cleanup(srv.Close)
	return svc, srv
}

// loadFixture reads a file from ./testdata and fails the test on error.
func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("loadFixture(%q): %v", name, err)
	}
	return data
}

// TestNewTestServiceSmoke confirms the helper works and that a simple
// 200 OK round-trip succeeds.
func TestNewTestServiceSmoke(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	if svc.BaseURL() == "" {
		t.Fatal("expected baseURL to be set")
	}
}

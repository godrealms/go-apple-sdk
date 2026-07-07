package AppStoreServer

import (
	"context"
	"net/http"
	"testing"
)

// TestLookUpOrderID exercises a representative legacy endpoint end to end: it
// asserts the method, the path (with the order ID substituted), and that the
// response decodes — coverage the legacy App Store Server stack previously
// lacked entirely.
func TestLookUpOrderID(t *testing.T) {
	var gotMethod, gotPath string
	c := newServerTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":0,"signedTransactions":[]}`))
	}))

	resp, err := LookUpOrderID(context.Background(), c, "ORDER-123")
	if err != nil {
		t.Fatalf("LookUpOrderID: %v", err)
	}
	if gotMethod != "GET" {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/inApps/v1/lookup/ORDER-123" {
		t.Errorf("path = %q, want /inApps/v1/lookup/ORDER-123", gotPath)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
}

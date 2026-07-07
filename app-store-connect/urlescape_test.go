package AppStoreConnect

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// TestService_PathSegmentEscaping guards the URL-injection bug: a caller-
// supplied ID containing URL metacharacters must be percent-encoded into the
// path, never interpreted as query parameters or a fragment.
func TestService_PathSegmentEscaping(t *testing.T) {
	var gotPath, gotRawQuery string
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"type":"apps","id":"x","attributes":{}}}`))
	}))

	// The "?injected=1" must not become a real query parameter.
	_, _ = svc.Apps().Get(context.Background(), "123?injected=1", nil)

	if strings.Contains(gotRawQuery, "injected") {
		t.Errorf("injected query parameter leaked from ID: rawQuery=%q", gotRawQuery)
	}
	if !strings.Contains(gotPath, "%3F") {
		t.Errorf("'?' in ID should be percent-encoded into the path, got path=%q", gotPath)
	}
}

package AppStoreConnect

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// captureLogger records every LogRecord it receives. Safe for
// concurrent use so it can also exercise the shared-service case.
type captureLogger struct {
	mu      sync.Mutex
	records []LogRecord
}

func (c *captureLogger) Log(r LogRecord) {
	c.mu.Lock()
	c.records = append(c.records, r)
	c.mu.Unlock()
}

func (c *captureLogger) snapshot() []LogRecord {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]LogRecord, len(c.records))
	copy(out, c.records)
	return out
}

func TestLogger_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "apps_list_page1.json"))
	}))
	defer srv.Close()

	logger := &captureLogger{}
	svc := New(Config{
		BaseURL:    srv.URL,
		Authorizer: noopAuthorizer,
		Logger:     logger,
	})
	if _, err := svc.Apps().List(context.Background(), nil); err != nil {
		t.Fatalf("List: %v", err)
	}
	recs := logger.snapshot()
	if len(recs) != 1 {
		t.Fatalf("records = %d, want 1", len(recs))
	}
	r := recs[0]
	if r.Method != "GET" {
		t.Errorf("method = %q", r.Method)
	}
	if r.StatusCode != 200 {
		t.Errorf("status = %d", r.StatusCode)
	}
	if r.Duration <= 0 {
		t.Errorf("duration = %v", r.Duration)
	}
	if r.Err != nil {
		t.Errorf("err = %v", r.Err)
	}
}

func TestLogger_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(loadFixture(t, "error_403.json"))
	}))
	defer srv.Close()

	logger := &captureLogger{}
	svc := New(Config{BaseURL: srv.URL, Authorizer: noopAuthorizer, Logger: logger})
	_, err := svc.Apps().List(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from 403 response")
	}

	recs := logger.snapshot()
	if len(recs) != 1 {
		t.Fatalf("records = %d", len(recs))
	}
	r := recs[0]
	if r.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d", r.StatusCode)
	}
	// Err field must carry the *APIError so callers can triage in one
	// place.
	var apiErr *APIError
	if !errors.As(r.Err, &apiErr) {
		t.Errorf("Err is not APIError: %T", r.Err)
	}
}

func TestLoggerFunc_Adapter(t *testing.T) {
	var called bool
	f := LoggerFunc(func(_ LogRecord) { called = true })
	f.Log(LogRecord{Method: "X"})
	if !called {
		t.Error("LoggerFunc did not forward Log")
	}
}

// TestLogger_Concurrent validates the thread-safety claim in the
// Logger docstring: a single Service may be shared across goroutines,
// so Logger.Log must be called without racing. Run under -race to
// actually detect a violation.
func TestLogger_Concurrent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(loadFixture(t, "apps_list_page1.json"))
	}))
	defer srv.Close()

	logger := &captureLogger{}
	svc := New(Config{BaseURL: srv.URL, Authorizer: noopAuthorizer, Logger: logger})

	const N = 16
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			if _, err := svc.Apps().List(context.Background(), nil); err != nil {
				t.Errorf("List: %v", err)
			}
		}()
	}
	wg.Wait()

	recs := logger.snapshot()
	if len(recs) != N {
		t.Fatalf("records = %d, want %d", len(recs), N)
	}
	for i, r := range recs {
		if r.Method != "GET" || r.StatusCode != 200 || r.Err != nil {
			t.Errorf("record %d: %+v", i, r)
		}
	}
}

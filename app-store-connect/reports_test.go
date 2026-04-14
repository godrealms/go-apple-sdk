package AppStoreConnect

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// gzipBytes returns a gzip stream containing payload, as the test
// stand-in for Apple's TSV report body.
func gzipBytes(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(payload); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func TestReports_DownloadSalesReport(t *testing.T) {
	wantTSV := []byte("Provider\tVendor\tUnits\nAPPLE\t85000001\t42\n")
	var captured struct {
		method string
		path   string
		query  string
		accept string
	}
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.query = r.URL.RawQuery
		captured.accept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/a-gzip")
		_, _ = w.Write(gzipBytes(t, wantTSV))
	}))

	body, err := svc.Reports().DownloadSalesReport(context.Background(), SalesReportRequest{
		VendorNumber:  "85000001",
		ReportType:    "SALES",
		ReportSubType: "SUMMARY",
		Frequency:     "DAILY",
		ReportDate:    "2025-03-01",
		Version:       "1_0",
	})
	if err != nil {
		t.Fatalf("DownloadSalesReport: %v", err)
	}

	if captured.method != "GET" {
		t.Errorf("method = %q", captured.method)
	}
	if captured.path != "/v1/salesReports" {
		t.Errorf("path = %q", captured.path)
	}
	if !strings.Contains(captured.accept, "application/a-gzip") {
		t.Errorf("Accept header missing a-gzip: %q", captured.accept)
	}
	for _, want := range []string{
		"filter%5BvendorNumber%5D=85000001",
		"filter%5BreportType%5D=SALES",
		"filter%5BreportSubType%5D=SUMMARY",
		"filter%5Bfrequency%5D=DAILY",
		"filter%5BreportDate%5D=2025-03-01",
		"filter%5Bversion%5D=1_0",
	} {
		if !strings.Contains(captured.query, want) {
			t.Errorf("query missing %q: %q", want, captured.query)
		}
	}
	if !bytes.Equal(body, wantTSV) {
		t.Errorf("body = %q, want %q", body, wantTSV)
	}
}

func TestReports_DownloadSalesReport_MissingFieldsLocalError(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	_, err := svc.Reports().DownloadSalesReport(context.Background(), SalesReportRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *ClientError
	if !errors.As(err, &ce) {
		t.Errorf("expected *ClientError, got %T: %v", err, err)
	}
}

func TestReports_DownloadSalesReport_APIError(t *testing.T) {
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(loadFixture(t, "error_403.json"))
	}))

	_, err := svc.Reports().DownloadSalesReport(context.Background(), SalesReportRequest{
		VendorNumber:  "85000001",
		ReportType:    "SALES",
		ReportSubType: "SUMMARY",
		Frequency:     "DAILY",
		ReportDate:    "2025-03-01",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}

func TestReports_DownloadSalesReport_GzipWithoutContentType(t *testing.T) {
	// Some middleboxes strip Content-Type. The sniff path must still
	// detect the gzip magic number and decode the body correctly.
	wantTSV := []byte("col1\tcol2\nA\tB\n")
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deliberately omit Content-Type to exercise the magic-number sniff.
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(gzipBytes(t, wantTSV))
	}))

	body, err := svc.Reports().DownloadSalesReport(context.Background(), SalesReportRequest{
		VendorNumber:  "85000001",
		ReportType:    "SALES",
		ReportSubType: "SUMMARY",
		Frequency:     "DAILY",
		ReportDate:    "2025-03-01",
	})
	if err != nil {
		t.Fatalf("DownloadSalesReport: %v", err)
	}
	if !bytes.Equal(body, wantTSV) {
		t.Errorf("body = %q, want %q", body, wantTSV)
	}
}

func TestReports_DownloadFinanceReport(t *testing.T) {
	wantTSV := []byte("Start_Date\tEnd_Date\tRegion\n2025-03-01\t2025-03-31\tUS\n")
	var capturedQuery string
	svc, _ := newTestService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/financeReports" {
			t.Errorf("path = %q", r.URL.Path)
		}
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/a-gzip")
		_, _ = w.Write(gzipBytes(t, wantTSV))
	}))

	body, err := svc.Reports().DownloadFinanceReport(context.Background(), FinanceReportRequest{
		VendorNumber: "85000001",
		RegionCode:   "US",
		ReportDate:   "2025-03",
		ReportType:   "FINANCIAL",
	})
	if err != nil {
		t.Fatalf("DownloadFinanceReport: %v", err)
	}
	if !bytes.Equal(body, wantTSV) {
		t.Errorf("body = %q, want %q", body, wantTSV)
	}
	for _, want := range []string{
		"filter%5BvendorNumber%5D=85000001",
		"filter%5BregionCode%5D=US",
		"filter%5BreportDate%5D=2025-03",
		"filter%5BreportType%5D=FINANCIAL",
	} {
		if !strings.Contains(capturedQuery, want) {
			t.Errorf("query missing %q: %q", want, capturedQuery)
		}
	}
}

func TestReports_DownloadFinanceReport_MissingFields(t *testing.T) {
	svc := New(Config{BaseURL: "http://example", Authorizer: noopAuthorizer})
	_, err := svc.Reports().DownloadFinanceReport(context.Background(), FinanceReportRequest{
		VendorNumber: "85000001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var ce *ClientError
	if !errors.As(err, &ce) {
		t.Errorf("expected *ClientError, got %T: %v", err, err)
	}
}

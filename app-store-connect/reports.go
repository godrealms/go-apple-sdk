package AppStoreConnect

import (
	"context"
	"fmt"
)

// ReportsService provides access to the sales and finance report
// endpoints on App Store Connect.
//
// Apple serves these reports as gzipped TSV files, keyed by a set of
// mandatory filter parameters (vendorNumber, reportType, frequency,
// reportDate, ...). Callers are responsible for supplying the correct
// combination of filters for the report they want.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/download_sales_and_trends_reports
// and https://developer.apple.com/documentation/appstoreconnectapi/download_finance_reports
type ReportsService struct {
	svc *Service
}

// SalesReportRequest describes a request for Apple's sales and trends
// report download endpoint.
//
// VendorNumber, ReportType, ReportSubType, Frequency, and ReportDate
// are all required by Apple — leaving any of them empty will make the
// request fail server-side with a 400. Version is optional and
// defaults (per Apple) to the current schema version when unset.
//
// Reference for valid values:
// https://developer.apple.com/help/app-store-connect/reference/report-types
type SalesReportRequest struct {
	VendorNumber  string // e.g. "85000001"
	ReportType    string // e.g. "SALES", "SUBSCRIBER"
	ReportSubType string // e.g. "SUMMARY", "DETAILED"
	Frequency     string // e.g. "DAILY", "WEEKLY", "MONTHLY", "YEARLY"
	ReportDate    string // e.g. "2025-03-01" (format depends on Frequency)
	Version       string // optional, e.g. "1_0"
}

// FinanceReportRequest describes a request for Apple's finance report
// download endpoint.
//
// RegionCode and ReportDate are required; ReportType defaults to
// "FINANCIAL" server-side but may also be set to "FINANCE_DETAIL".
//
// Reference: https://developer.apple.com/documentation/appstoreconnectapi/download_finance_reports
type FinanceReportRequest struct {
	VendorNumber string
	RegionCode   string // e.g. "US", "Z1" (rest of world)
	ReportDate   string // e.g. "2025-03" for monthly financial
	ReportType   string // optional, defaults to "FINANCIAL"
}

// DownloadSalesReport requests a sales report and returns its
// decompressed TSV bytes. The returned slice is the raw TSV as Apple
// serves it — the first line contains the column headers.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/download_sales_and_trends_reports
func (s *ReportsService) DownloadSalesReport(ctx context.Context, req SalesReportRequest) ([]byte, error) {
	if req.VendorNumber == "" {
		return nil, &ClientError{Message: "Reports.DownloadSalesReport: VendorNumber is required"}
	}
	if req.ReportType == "" || req.ReportSubType == "" || req.Frequency == "" || req.ReportDate == "" {
		return nil, &ClientError{Message: "Reports.DownloadSalesReport: ReportType, ReportSubType, Frequency, and ReportDate are required"}
	}
	q := NewQuery().
		Filter("vendorNumber", req.VendorNumber).
		Filter("reportType", req.ReportType).
		Filter("reportSubType", req.ReportSubType).
		Filter("frequency", req.Frequency).
		Filter("reportDate", req.ReportDate)
	if req.Version != "" {
		q = q.Filter("version", req.Version)
	}
	_, body, err := s.svc.doRaw(ctx, "GET", "/v1/salesReports", q)
	if err != nil {
		return nil, fmt.Errorf("download sales report: %w", err)
	}
	return body, nil
}

// DownloadFinanceReport requests a finance report and returns its
// decompressed TSV bytes.
//
// See https://developer.apple.com/documentation/appstoreconnectapi/download_finance_reports
func (s *ReportsService) DownloadFinanceReport(ctx context.Context, req FinanceReportRequest) ([]byte, error) {
	if req.VendorNumber == "" {
		return nil, &ClientError{Message: "Reports.DownloadFinanceReport: VendorNumber is required"}
	}
	if req.RegionCode == "" || req.ReportDate == "" {
		return nil, &ClientError{Message: "Reports.DownloadFinanceReport: RegionCode and ReportDate are required"}
	}
	q := NewQuery().
		Filter("vendorNumber", req.VendorNumber).
		Filter("regionCode", req.RegionCode).
		Filter("reportDate", req.ReportDate)
	if req.ReportType != "" {
		q = q.Filter("reportType", req.ReportType)
	}
	_, body, err := s.svc.doRaw(ctx, "GET", "/v1/financeReports", q)
	if err != nil {
		return nil, fmt.Errorf("download finance report: %w", err)
	}
	return body, nil
}

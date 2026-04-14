// Example: download a daily sales report (gzipped TSV) and print
// the column headers plus the first few rows.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	Apple "github.com/godrealms/go-apple-sdk"
	AppStoreConnect "github.com/godrealms/go-apple-sdk/app-store-connect"
)

func main() {
	kid := ""           // Your private key ID
	iss := ""           // Your issuer ID
	bid := ""           // Your app's primary bundle ID
	privateKey := ""    // Your ES256 private key (PKCS#8)
	vendorNumber := "" // Your vendor number (see Payments and Financial Reports)
	reportDate := "2025-03-01"

	client := Apple.NewClient(false, kid, iss, bid, privateKey)
	svc := client.AppStoreConnect()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	body, err := svc.Reports().DownloadSalesReport(ctx, AppStoreConnect.SalesReportRequest{
		VendorNumber:  vendorNumber,
		ReportType:    "SALES",
		ReportSubType: "SUMMARY",
		Frequency:     "DAILY",
		ReportDate:    reportDate,
		Version:       "1_0",
	})
	if err != nil {
		var apiErr *AppStoreConnect.APIError
		if errors.As(err, &apiErr) {
			log.Fatalf("Apple rejected the request: HTTP %d, %d errors", apiErr.StatusCode, len(apiErr.Errors))
		}
		log.Fatalf("local error: %v", err)
	}

	// Persist the full TSV so callers can process it with the data
	// pipeline of their choice, then dump the first 5 lines to stdout
	// as a quick sanity check.
	outPath := fmt.Sprintf("sales-%s.tsv", reportDate)
	if err := os.WriteFile(outPath, body, 0o644); err != nil {
		log.Fatalf("write %s: %v", outPath, err)
	}
	log.Printf("wrote %d bytes to %s", len(body), outPath)

	for i, line := range bytes.SplitN(body, []byte{'\n'}, 6) {
		if i >= 5 {
			break
		}
		fmt.Printf("%d: %s\n", i, line)
	}
}

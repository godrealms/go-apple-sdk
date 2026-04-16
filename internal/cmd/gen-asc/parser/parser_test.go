package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/ir"
)

func TestParse_MinimalFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "minimal_spec.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	doc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// Metadata is populated by main.go in Task 10; Parse itself only
	// deals with spec contents.

	if got, want := len(doc.Resources), 2; got != want {
		t.Fatalf("resources = %d, want %d (names: %v)", got, want, resourceNames(doc.Resources))
	}

	apps := findResource(t, doc.Resources, "apps")
	if got, want := apps.Name, "Apps"; got != want {
		t.Errorf("apps.Name = %q, want %q", got, want)
	}
	if got, want := len(apps.Operations), 5; got != want {
		t.Errorf("apps ops = %d, want %d", got, want)
	}

	// Assert each operation's name/verb/path so a regression in
	// operationName cannot silently pass.
	wantOps := []struct {
		Name, HTTPMethod, PathTemplate string
	}{
		{"List", "GET", "/v1/apps"},
		{"Delete", "DELETE", "/v1/apps/{id}"},
		{"Get", "GET", "/v1/apps/{id}"},
		{"Update", "PATCH", "/v1/apps/{id}"},
		{"GetRelationshipsAppInfos", "GET", "/v1/apps/{id}/relationships/appInfos"},
	}
	if len(apps.Operations) == len(wantOps) {
		for i, w := range wantOps {
			got := apps.Operations[i]
			if got.Name != w.Name || got.HTTPMethod != w.HTTPMethod || got.PathTemplate != w.PathTemplate {
				t.Errorf("apps.Operations[%d] = (%q,%q,%q), want (%q,%q,%q)",
					i, got.Name, got.HTTPMethod, got.PathTemplate, w.Name, w.HTTPMethod, w.PathTemplate)
			}
		}
	}

	reports := findResource(t, doc.Resources, "analyticsReports")
	if got, want := reports.Name, "AnalyticsReports"; got != want {
		t.Errorf("reports.Name = %q, want %q", got, want)
	}
	if got, want := len(reports.Operations), 1; got != want {
		t.Errorf("reports ops = %d, want %d", got, want)
	}
}

// helpers

func findResource(t *testing.T, rs []ir.Resource, apiName string) ir.Resource {
	t.Helper()
	for _, r := range rs {
		if r.APIName == apiName {
			return r
		}
	}
	t.Fatalf("resource %q not found, have %v", apiName, resourceNames(rs))
	return ir.Resource{}
}

func resourceNames(rs []ir.Resource) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.APIName
	}
	return out
}

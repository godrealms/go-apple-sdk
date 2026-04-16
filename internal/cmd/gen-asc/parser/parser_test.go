package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/ir"
	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/skip"
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

// TestParse_SkipsHandWrittenResources verifies the WithSkipSet
// option drops resources whose API name appears in the skip set.
func TestParse_SkipsHandWrittenResources(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "minimal_spec.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	skipSet, err := skip.Load(filepath.Join("testdata", "skip_apps.txt"))
	if err != nil {
		t.Fatalf("skip.Load: %v", err)
	}
	doc, err := Parse(data, WithSkipSet(skipSet))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// "apps" should be filtered; only "analyticsReports" survives.
	if got, want := len(doc.Resources), 1; got != want {
		t.Fatalf("resources = %d, want %d (names: %v)", got, want, resourceNames(doc.Resources))
	}
	if doc.Resources[0].APIName != "analyticsReports" {
		t.Errorf("surviving resource = %q, want analyticsReports", doc.Resources[0].APIName)
	}
}

// TestParse_NilSkipSet confirms WithSkipSet(nil) is a no-op so
// callers can pass conditional options without branching.
func TestParse_NilSkipSet(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "minimal_spec.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	doc, err := Parse(data, WithSkipSet(nil))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got, want := len(doc.Resources), 2; got != want {
		t.Errorf("resources = %d, want %d (nil skip set should not filter)", got, want)
	}
}

// TestParse_MissingOpenAPI exercises the decodeRaw guard for a spec
// that lacks the top-level "openapi" key.
func TestParse_MissingOpenAPI(t *testing.T) {
	_, err := Parse([]byte(`{"paths": {"/v1/foo": {}}}`))
	if err == nil {
		t.Fatal("expected error for missing openapi version")
	}
	if !strings.Contains(err.Error(), "openapi") {
		t.Errorf("err = %v, want it to mention 'openapi'", err)
	}
}

// TestParse_NoPaths exercises the decodeRaw guard for a spec with
// an empty paths map.
func TestParse_NoPaths(t *testing.T) {
	_, err := Parse([]byte(`{"openapi": "3.0.1", "paths": {}}`))
	if err == nil {
		t.Fatal("expected error for empty paths")
	}
	if !strings.Contains(err.Error(), "paths") {
		t.Errorf("err = %v, want it to mention 'paths'", err)
	}
}

// TestParse_MalformedJSON triggers the json.Unmarshal failure path
// in decodeRaw.
func TestParse_MalformedJSON(t *testing.T) {
	_, err := Parse([]byte(`{not json`))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

// TestOperationName_VerbCoverage exercises every branch of
// operationName + verbPrefix on a synthetic resource so a
// regression in the canonical-name table cannot pass silently.
func TestOperationName_VerbCoverage(t *testing.T) {
	cases := []struct {
		path, verb, want string
	}{
		// Canonical fast paths
		{"/v1/widgets", "GET", "List"},
		{"/v1/widgets", "POST", "Create"},
		{"/v1/widgets/{id}", "GET", "Get"},
		{"/v1/widgets/{id}", "PATCH", "Update"},
		{"/v1/widgets/{id}", "PUT", "Replace"},
		{"/v1/widgets/{id}", "DELETE", "Delete"},
		// Fallback paths exercise verbPrefix for every verb
		{"/v1/widgets/{id}/actions/refresh", "GET", "GetActionsRefresh"},
		{"/v1/widgets/{id}/actions/refresh", "POST", "CreateActionsRefresh"},
		{"/v1/widgets/{id}/actions/refresh", "PATCH", "UpdateActionsRefresh"},
		{"/v1/widgets/{id}/actions/refresh", "PUT", "ReplaceActionsRefresh"},
		{"/v1/widgets/{id}/actions/refresh", "DELETE", "DeleteActionsRefresh"},
		// verbPrefix default branch (unknown verb)
		{"/v1/widgets/{id}/actions/refresh", "OPTIONS", "OptionsActionsRefresh"},
	}
	for _, c := range cases {
		got := operationName("widgets", pathOp{Path: c.path, Verb: c.verb, Op: &rawOperation{}})
		if got != c.want {
			t.Errorf("operationName(%q, %q) = %q, want %q", c.path, c.verb, got, c.want)
		}
	}
}

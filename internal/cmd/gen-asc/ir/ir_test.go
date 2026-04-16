package ir

import (
	"bytes"
	"encoding/json"
	"testing"
)

// TestDocument_JSONRoundTrip protects against tag typos: if every
// field survives a Marshal → Unmarshal cycle without data loss, the
// JSON tags are at least internally consistent.
func TestDocument_JSONRoundTrip(t *testing.T) {
	orig := Document{
		Metadata: Metadata{
			SpecVersion: "3.7.1",
			SpecSHA256:  "deadbeef",
			SpecSource:  "https://example.test/spec.zip",
			GeneratedAt: "2026-04-16T00:00:00Z",
		},
		Resources: []Resource{
			{
				Name:    "AnalyticsReports",
				APIName: "analyticsReports",
				Operations: []Operation{
					{
						Name:         "List",
						HTTPMethod:   "GET",
						PathTemplate: "/v1/analyticsReports",
						QueryParams: []Field{
							{Name: "Filter", JSONName: "filter[category]", GoType: "string"},
						},
					},
				},
				Attrs: []Field{
					{Name: "Category", JSONName: "category", GoType: "string", Required: true},
				},
				Rels: []Relationship{
					{Name: "App", Target: "apps"},
				},
			},
		},
	}

	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var round Document
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	b2, err := json.Marshal(round)
	if err != nil {
		t.Fatalf("marshal round-trip: %v", err)
	}
	if !bytes.Equal(b, b2) {
		t.Errorf("round trip changed bytes\n  first: %s\n  second: %s", b, b2)
	}
}

// TestDocument_JSONRoundTrip_AllFields complements the happy-path
// round-trip by populating every omitempty field with a non-zero
// value. Without this, a typo on any omitempty tag would silently
// pass the original test because the field would never appear in
// the serialized bytes.
func TestDocument_JSONRoundTrip_AllFields(t *testing.T) {
	orig := Document{
		Metadata: Metadata{
			SpecVersion: "3.7.1",
			SpecSHA256:  "deadbeef",
			SpecSource:  "https://example.test/spec.zip",
			GeneratedAt: "2026-04-16T00:00:00Z",
		},
		Resources: []Resource{
			{
				Name:    "AnalyticsReports",
				APIName: "analyticsReports",
				DocURL:  "https://developer.apple.com/documentation/appstoreconnectapi/analyticsreports",
				Operations: []Operation{
					{
						Name:         "Create",
						HTTPMethod:   "POST",
						PathTemplate: "/v1/analyticsReports/{id}",
						PathParams: []Field{
							{Name: "ID", JSONName: "id", GoType: "string", Required: true, Comment: "report id"},
						},
						QueryParams: []Field{
							{Name: "Filter", JSONName: "filter[category]", GoType: "string", Comment: "category filter"},
						},
						RequestBody: &Type{
							Name: "AnalyticsReportCreateRequest",
							Fields: []Field{
								{Name: "Category", JSONName: "category", GoType: "string", Required: true, Comment: "report category"},
							},
						},
						ResponseBody: &Type{
							Name: "AnalyticsReportResponse",
							Fields: []Field{
								{Name: "ID", JSONName: "id", GoType: "string", Required: true, Comment: "created id"},
							},
						},
						DocURL:     "https://developer.apple.com/documentation/appstoreconnectapi/create_an_analytics_report",
						Deprecated: true,
						Summary:    "Create a new analytics report",
					},
				},
				Attrs: []Field{
					{Name: "Category", JSONName: "category", GoType: "string", Required: true, Comment: "report category"},
				},
				Rels: []Relationship{
					{Name: "App", Target: "apps", ToMany: true},
				},
			},
		},
	}

	b, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round Document
	if err := json.Unmarshal(b, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	b2, err := json.Marshal(round)
	if err != nil {
		t.Fatalf("marshal round-trip: %v", err)
	}
	if !bytes.Equal(b, b2) {
		t.Errorf("round trip changed bytes\n  first: %s\n  second: %s", b, b2)
	}
}

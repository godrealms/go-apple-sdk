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

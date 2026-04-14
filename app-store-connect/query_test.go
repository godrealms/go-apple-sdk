package AppStoreConnect

import (
	"net/url"
	"testing"
)

func TestQuery_EncodeEmpty(t *testing.T) {
	if got := (*Query)(nil).Encode(); got != "" {
		t.Errorf("nil query encode = %q, want empty", got)
	}
	if got := NewQuery().Encode(); got != "" {
		t.Errorf("empty query encode = %q, want empty", got)
	}
}

func TestQuery_EncodeAllPrimitives(t *testing.T) {
	q := NewQuery().
		Filter("bundleId", "com.acme.widgets").
		Filter("name", "Acme", "Gadgets").
		Fields("apps", "name", "bundleId").
		Include("appStoreVersions", "ciProduct").
		Sort("-updatedAt", "name").
		Limit(200).
		Cursor("CURSOR_XYZ").
		Set("foo", "bar")

	encoded := q.Encode()
	// Decode to a map so we don't depend on key ordering (url.Values.Encode
	// already sorts alphabetically, but parsing is the robust check).
	vals, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatalf("ParseQuery(%q) err: %v", encoded, err)
	}

	cases := map[string]string{
		"filter[bundleId]": "com.acme.widgets",
		"filter[name]":     "Acme,Gadgets",
		"fields[apps]":     "name,bundleId",
		"include":          "appStoreVersions,ciProduct",
		"sort":             "-updatedAt,name",
		"limit":            "200",
		"cursor":           "CURSOR_XYZ",
		"foo":              "bar",
	}
	for k, want := range cases {
		if got := vals.Get(k); got != want {
			t.Errorf("query[%s] = %q, want %q", k, got, want)
		}
	}
}

func TestQuery_EncodeDeterministic(t *testing.T) {
	// Build the same logical query two different ways and confirm that
	// the encoded output is byte-identical.
	a := NewQuery().Filter("b", "2").Filter("a", "1").Fields("apps", "z", "y")
	b := NewQuery().Filter("a", "1").Filter("b", "2").Fields("apps", "z", "y")
	if a.Encode() != b.Encode() {
		t.Errorf("non-deterministic encode:\n  a=%q\n  b=%q", a.Encode(), b.Encode())
	}
}

func TestQuery_LimitIgnoresNonPositive(t *testing.T) {
	q := NewQuery().Limit(0)
	if got := q.Encode(); got != "" {
		t.Errorf("Limit(0) encoded to %q, want empty", got)
	}
	q = NewQuery().Limit(-5)
	if got := q.Encode(); got != "" {
		t.Errorf("Limit(-5) encoded to %q, want empty", got)
	}
}

func TestQuery_Clone(t *testing.T) {
	original := NewQuery().
		Filter("bundleId", "com.acme.widgets").
		Include("appStoreVersions").
		Limit(50)
	clone := original.Clone()

	// Mutate the clone; original must be unchanged.
	clone.Filter("bundleId", "com.acme.gadgets").Limit(100)

	if want := "filter%5BbundleId%5D=com.acme.widgets&include=appStoreVersions&limit=50"; original.Encode() != want {
		t.Errorf("original mutated:\n  got:  %q\n  want: %q", original.Encode(), want)
	}
	if !contains(clone.Encode(), "limit=100") {
		t.Errorf("clone missing limit=100: %q", clone.Encode())
	}
}

func TestQuery_NilCloneIsEmpty(t *testing.T) {
	var q *Query
	if q.Clone().Encode() != "" {
		t.Errorf("nil clone should be empty")
	}
}

// contains is a tiny local helper to avoid pulling in strings just for
// two substring assertions.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

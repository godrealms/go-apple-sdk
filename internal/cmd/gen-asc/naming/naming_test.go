package naming

import "testing"

func TestPascalCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// Pure camelCase input (OpenAPI resource names are always this form)
		{"apps", "Apps"},
		{"analyticsReports", "AnalyticsReports"},
		{"appStoreVersions", "AppStoreVersions"},
		{"customerReviewResponses", "CustomerReviewResponses"},
		// Edge: already PascalCase
		{"App", "App"},
		// Edge: empty
		{"", ""},
		// Edge: single char
		{"a", "A"},
	}
	for _, c := range cases {
		if got := PascalCase(c.in); got != c.want {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCamelCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"AnalyticsReports", "analyticsReports"},
		{"App", "app"},
		{"", ""},
	}
	for _, c := range cases {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

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
		{"A", "a"},
		{"", ""},
	}
	for _, c := range cases {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPascalCase_Acronyms(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"bundleId", "BundleID"},
		{"appId", "AppID"},
		{"purchaseUrl", "PurchaseURL"},
		{"subscriptionStatusUrl", "SubscriptionStatusURL"},
		{"apiKey", "APIKey"},
		{"jsonWebToken", "JSONWebToken"},
		{"isJwtExpired", "IsJWTExpired"},
		{"jwtToken", "JWTToken"},
		// Acronym at end-of-word only: should NOT eat substrings of
		// longer words. "widget" must stay "Widget", not "WIDGetIDen".
		{"widget", "Widget"},
		{"idle", "Idle"},
		// False-positive guards: English words that END in acronym
		// characters must NOT be rewritten, because the split-then-
		// exact-match rule only touches isolated words.
		{"rapid", "Rapid"},
		{"valid", "Valid"},
		{"android", "Android"},
		{"squid", "Squid"},
		// Acronym at head of compound: "apiKey" -> ["Api","Key"]
		// -> "API" + "Key" -> "APIKey" (already covered, but make
		// the head-position intent explicit).
		{"apiResponse", "APIResponse"},
	}
	for _, c := range cases {
		if got := PascalCase(c.in); got != c.want {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestAcronymsLongestFirst guards the longest-first ordering
// invariant documented on the acronyms var. If someone adds a new
// acronym out of order (e.g., "JS" after "JSON"), this test fails
// loudly instead of silently producing wrong output.
func TestAcronymsLongestFirst(t *testing.T) {
	for i := 1; i < len(acronyms); i++ {
		if len(acronyms[i]) > len(acronyms[i-1]) {
			t.Errorf("acronyms[%d]=%q is longer than acronyms[%d]=%q; list must be longest-first",
				i, acronyms[i], i-1, acronyms[i-1])
		}
	}
}

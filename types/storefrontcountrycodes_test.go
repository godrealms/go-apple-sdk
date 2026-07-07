package types

import "testing"

// TestStorefrontCountryCode_IsValid guards against the reversed-logic bug:
// a valid storefront code is an ISO 3166-1 Alpha-3 (three-letter) code, not
// the empty string.
func TestStorefrontCountryCode_IsValid(t *testing.T) {
	cases := []struct {
		in   StorefrontCountryCode
		want bool
	}{
		{"USA", true},
		{"CHN", true},
		{"", false},
		{"US", false},   // two letters, not Alpha-3
		{"USAX", false}, // four letters
	}
	for _, tc := range cases {
		if got := tc.in.IsValid(); got != tc.want {
			t.Errorf("IsValid(%q) = %v, want %v", string(tc.in), got, tc.want)
		}
	}
}

// Package naming converts OpenAPI identifier strings (lowerCamelCase)
// to Go identifiers (UpperCamelCase / PascalCase), preserving a
// curated set of acronyms (URL, ID, API, JWT, JSON) in all-caps.
//
// All functions are pure, safe for concurrent use, and deterministic.
package naming

import "unicode"

// PascalCase converts a lowerCamelCase identifier such as
// "analyticsReports" into "AnalyticsReports". It is the identity
// function on already-PascalCase input. Empty input returns empty
// output.
//
// Acronym handling (URL / ID / …) lives in applyAcronyms and is
// applied as a post-processing step so the rules stay in one place.
func PascalCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return applyAcronyms(string(runes))
}

// CamelCase is the inverse of PascalCase: "AnalyticsReports" →
// "analyticsReports". Used when the generator needs to emit the
// JSON-tag value from a Go field name.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

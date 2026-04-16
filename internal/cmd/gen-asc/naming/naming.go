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
// Input contract: s must be lowerCamelCase. Inputs containing "-" or "_"
// are out of scope and will produce invalid Go identifiers.
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

// CamelCase lowercases the first rune of s, turning Go PascalCase
// identifiers such as "AnalyticsReports" into JSON-tag values such
// as "analyticsReports". It is NOT the inverse of PascalCase once
// Task 3's acronym carve-outs are active: CamelCase(PascalCase("appId"))
// is not guaranteed to equal "appId". Use it only for emitting JSON
// tags from a Go field name you already know is safe.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

package naming

import (
	"strings"
	"unicode"
)

// acronyms lists tokens that must stay fully uppercase in Go
// identifiers. They are matched at word boundaries: a "word" starts
// at a PascalCase capital or the beginning of the string.
//
// Order matters: longer acronyms are matched first so "JSON" takes
// precedence over "JS". Keep this list sorted from longest to
// shortest.
var acronyms = []string{
	"JSON",
	"JWT",
	"API",
	"URL",
	"ID",
}

// applyAcronyms upper-cases any word whose entire content equals one
// of the listed acronyms. It scans word-by-word (word = leading
// capital + following lowercase letters/digits) and rewrites each
// word in isolation.
//
// Examples (inputs are already PascalCase because PascalCase runs
// first):
//
//	"BundleId"              -> "BundleID"
//	"SubscriptionStatusUrl" -> "SubscriptionStatusURL"
//	"IsJwtExpired"          -> "IsJWTExpired"
//	"Rapid"                 -> "Rapid"  (no exact-match word)
//	"Widget"                -> "Widget"
func applyAcronyms(s string) string {
	if s == "" {
		return s
	}
	words := splitPascalWords(s)
	for i, w := range words {
		words[i] = rewriteTailAcronym(w)
	}
	return strings.Join(words, "")
}

// splitPascalWords splits "SubscriptionStatusUrl" into
// ["Subscription","Status","Url"]. A word boundary is an uppercase
// rune that follows a lowercase rune (or is the first rune).
func splitPascalWords(s string) []string {
	if s == "" {
		return nil
	}
	runes := []rune(s)
	var out []string
	start := 0
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) && unicode.IsLower(runes[i-1]) {
			out = append(out, string(runes[start:i]))
			start = i
		}
	}
	out = append(out, string(runes[start:]))
	return out
}

// rewriteTailAcronym upper-cases a word if and only if the entire
// word equals one of the listed acronyms (case-insensitively).
// splitPascalWords already isolates words at PascalCase boundaries,
// so an acronym must BE a word, not just end one — otherwise
// "rapid" would become "RapID" because "id" is a suffix of "rapid".
//
//	"Url"    -> "URL"   (exact match)
//	"Id"     -> "ID"    (exact match)
//	"Rapid"  -> "Rapid" (no exact match; "id" is a suffix, not a word)
//	"Widget" -> "Widget"
func rewriteTailAcronym(word string) string {
	lower := strings.ToLower(word)
	for _, a := range acronyms {
		if lower == strings.ToLower(a) {
			return a
		}
	}
	return word
}

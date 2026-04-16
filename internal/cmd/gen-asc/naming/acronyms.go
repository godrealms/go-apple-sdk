package naming

import "strings"

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

// applyAcronyms upper-cases any trailing acronym in a PascalCase
// word. It scans word-by-word (word = leading capital + following
// lowercase letters/digits) and rewrites each word's tail if it ends
// with one of the listed acronyms in any case.
//
// Examples (inputs are already PascalCase because PascalCase runs
// first):
//
//	"BundleId"              -> "BundleID"
//	"SubscriptionStatusUrl" -> "SubscriptionStatusURL"
//	"IsJwtExpired"          -> "IsJWTExpired"
//	"Widget"                -> "Widget"  (no tail acronym)
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
		if isUpper(runes[i]) && isLower(runes[i-1]) {
			out = append(out, string(runes[start:i]))
			start = i
		}
	}
	out = append(out, string(runes[start:]))
	return out
}

// rewriteTailAcronym checks whether a single word ends with one of
// the acronyms (case-insensitive) and upper-cases that tail.
//
//	"Url"      -> "URL"
//	"AppId"    -> "AppID"
//	"Idle"     -> "Idle"  (trailing "le" doesn't match)
//	"Widget"   -> "Widget"
func rewriteTailAcronym(word string) string {
	lower := strings.ToLower(word)
	for _, a := range acronyms {
		la := strings.ToLower(a)
		if strings.HasSuffix(lower, la) && len(word) >= len(la) {
			head := word[:len(word)-len(la)]
			return head + a
		}
	}
	return word
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }

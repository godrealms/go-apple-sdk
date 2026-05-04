// Package skip loads the "hand-written services" exclusion list from
// a plain-text file. Two line formats are recognised:
//
//	resource                 — skip every operation on the resource
//	resource path-pattern    — skip operations on the resource whose
//	                           path template matches the pattern
//
// path-pattern is either an exact path template (e.g. "/v1/apps/{id}")
// or a prefix terminated by "/*" (e.g. "/v1/apps/{id}/relationships/*").
// Prefix matches accept the prefix itself plus anything below.
//
// Comments start with "#" and extend to end-of-line. Leading/trailing
// whitespace on each line is trimmed. Blank lines and pure-comment
// lines are skipped. Duplicate entries collapse silently.
//
// Why two granularities? Apple's spec aggregates many sub-paths under
// one resource (e.g. "apps" owns 86 operations including
// /v1/apps/{id}/relationships/*). The hand-written AppsService only
// covers /v1/apps and /v1/apps/{id}. Resource-level skip would drop
// the other 84 ops; per-path skip lets the generator produce them.
package skip

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Set is an immutable view over the loaded skip list. Zero value is
// an empty set (safe to call Contains / Skip on).
type Set struct {
	// resources skips every operation on the named resource.
	resources map[string]struct{}
	// paths is per-resource path patterns: each entry skips
	// operations whose path template matches the pattern.
	paths map[string][]pathPattern
}

type pathPattern struct {
	raw      string // original text (for debugging / round-trip)
	pattern  string // with trailing "/*" stripped if isPrefix
	isPrefix bool   // true when raw ended with "/*"
}

// Load reads and parses a skip-list file from disk.
func Load(path string) (*Set, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("skip: open %q: %w", path, err)
	}
	defer f.Close()

	s := &Set{
		resources: make(map[string]struct{}),
		paths:     make(map[string][]pathPattern),
	}
	sc := bufio.NewScanner(f)
	lineno := 0
	for sc.Scan() {
		lineno++
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		switch len(fields) {
		case 1:
			// Resource-level skip.
			s.resources[fields[0]] = struct{}{}
		case 2:
			// Per-path skip.
			resource, pat := fields[0], fields[1]
			pp := pathPattern{raw: pat}
			if strings.HasSuffix(pat, "/*") {
				pp.isPrefix = true
				pp.pattern = strings.TrimSuffix(pat, "/*")
			} else {
				pp.pattern = pat
			}
			s.paths[resource] = append(s.paths[resource], pp)
		default:
			return nil, fmt.Errorf("skip: %s:%d: expected 1 or 2 tokens, got %d (%q)",
				path, lineno, len(fields), line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("skip: scan %q: %w", path, err)
	}
	return s, nil
}

// Contains reports whether the given OpenAPI resource name is in the
// resource-level skip list. It does NOT consult per-path patterns —
// use Skip for finer per-operation filtering.
//
// Safe on a nil or zero-value *Set.
func (s *Set) Contains(name string) bool {
	if s == nil || s.resources == nil {
		return false
	}
	_, ok := s.resources[name]
	return ok
}

// Skip reports whether the operation at (resource, path) should be
// excluded from generation. Returns true if either:
//   - the resource is in the resource-level skip list (Contains), or
//   - the path matches one of the resource's per-path patterns.
//
// Safe on a nil or zero-value *Set.
func (s *Set) Skip(resource, path string) bool {
	if s == nil {
		return false
	}
	if _, ok := s.resources[resource]; ok {
		return true
	}
	for _, pp := range s.paths[resource] {
		if pp.isPrefix {
			if path == pp.pattern || strings.HasPrefix(path, pp.pattern+"/") {
				return true
			}
		} else {
			if path == pp.pattern {
				return true
			}
		}
	}
	return false
}

// Len returns the number of entries in the skip list (resource-level
// plus per-path).
func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	n := len(s.resources)
	for _, v := range s.paths {
		n += len(v)
	}
	return n
}

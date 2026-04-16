// Package skip loads the "hand-written services" exclusion list from
// a plain-text file. Each non-blank, non-comment line is treated as an
// OpenAPI resource name (lowerCamel). Comments start with "#" and
// extend to end-of-line — a "#" anywhere on a line begins a comment,
// so resource names may not contain "#" (Apple's lowerCamel schema
// never does). Leading/trailing whitespace on each line is trimmed.
// Duplicates are silently collapsed.
//
// The file format is deliberately minimal to avoid pulling in a YAML
// dependency for what is fundamentally a string list.
package skip

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Set is an immutable view over the loaded skip list. Zero value is
// an empty set (safe to call Contains on).
type Set struct {
	m map[string]struct{}
}

// Load reads and parses a skip-list file from disk.
func Load(path string) (*Set, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("skip: open %q: %w", path, err)
	}
	defer f.Close()

	m := make(map[string]struct{})
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m[line] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("skip: scan %q: %w", path, err)
	}
	return &Set{m: m}, nil
}

// Contains reports whether the given OpenAPI resource name is in the
// skip list. It is safe on a nil or zero-value *Set.
func (s *Set) Contains(name string) bool {
	if s == nil || s.m == nil {
		return false
	}
	_, ok := s.m[name]
	return ok
}

// Len returns the number of unique entries in the skip list.
func (s *Set) Len() int {
	if s == nil {
		return 0
	}
	return len(s.m)
}

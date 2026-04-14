package AppStoreConnect

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Query is a fluent builder for App Store Connect query parameters.
//
// It supports the JSON:API query primitives Apple documents:
//
//   - filter[<name>]=v1,v2     via [Query.Filter]
//   - fields[<type>]=f1,f2      via [Query.Fields]
//   - include=rel1,rel2         via [Query.Include]
//   - sort=-name,updatedAt      via [Query.Sort]
//   - limit=200                 via [Query.Limit]
//   - cursor=<opaque>           via [Query.Cursor]
//
// All methods return the receiver so calls can be chained. Query is
// designed to be built and then passed by pointer; once encoded it
// produces deterministic output regardless of insertion order of
// filters/fields so tests are stable.
//
// A nil *Query is a valid empty query and encodes to "".
type Query struct {
	filters  map[string][]string
	fields   map[string][]string
	include  []string
	sort     []string
	limit    int
	cursor   string
	extra    map[string][]string
	hasLimit bool
}

// NewQuery constructs an empty query.
func NewQuery() *Query {
	return &Query{
		filters: make(map[string][]string),
		fields:  make(map[string][]string),
		extra:   make(map[string][]string),
	}
}

// Filter adds a filter[<name>]=<values...> constraint. Calling Filter
// twice for the same name appends values (JSON:API convention: comma
// separated list).
func (q *Query) Filter(name string, values ...string) *Query {
	q.ensure()
	q.filters[name] = append(q.filters[name], values...)
	return q
}

// Fields restricts the attributes returned for a given resource type
// via fields[<resourceType>]=<fieldNames>.
func (q *Query) Fields(resourceType string, fieldNames ...string) *Query {
	q.ensure()
	q.fields[resourceType] = append(q.fields[resourceType], fieldNames...)
	return q
}

// Include requests that related resources of the given names be inlined
// in the "included" section of the response.
func (q *Query) Include(relationships ...string) *Query {
	q.ensure()
	q.include = append(q.include, relationships...)
	return q
}

// Sort sets the sort order. Each value should be a field name, optionally
// prefixed with "-" for descending order (e.g. "-updatedAt").
func (q *Query) Sort(sortFields ...string) *Query {
	q.ensure()
	q.sort = append(q.sort, sortFields...)
	return q
}

// Limit sets the page size (JSON:API limit). Apple caps individual
// endpoints at their own maximums; values <= 0 are treated as "unset".
func (q *Query) Limit(n int) *Query {
	q.ensure()
	if n > 0 {
		q.limit = n
		q.hasLimit = true
	}
	return q
}

// Cursor sets the opaque pagination cursor (extracted from links.next)
// to fetch the subsequent page.
func (q *Query) Cursor(cursor string) *Query {
	q.ensure()
	q.cursor = cursor
	return q
}

// Set adds a raw query parameter. This is an escape hatch for
// endpoint-specific parameters not covered by the structured builders.
func (q *Query) Set(key string, values ...string) *Query {
	q.ensure()
	q.extra[key] = append(q.extra[key], values...)
	return q
}

// Clone returns a deep copy of the query so callers can derive variants
// without mutating a shared base query.
func (q *Query) Clone() *Query {
	if q == nil {
		return NewQuery()
	}
	out := NewQuery()
	for k, v := range q.filters {
		out.filters[k] = append([]string(nil), v...)
	}
	for k, v := range q.fields {
		out.fields[k] = append([]string(nil), v...)
	}
	out.include = append([]string(nil), q.include...)
	out.sort = append([]string(nil), q.sort...)
	out.limit = q.limit
	out.hasLimit = q.hasLimit
	out.cursor = q.cursor
	for k, v := range q.extra {
		out.extra[k] = append([]string(nil), v...)
	}
	return out
}

// Encode renders the query as a URL-encoded query string (without the
// leading "?"). Keys are emitted in a deterministic, alphabetized order
// so callers can build stable URLs and reliably assert on them in tests.
//
// A nil receiver encodes to "".
func (q *Query) Encode() string {
	if q == nil {
		return ""
	}
	v := url.Values{}
	for name, values := range q.filters {
		if len(values) > 0 {
			v.Set("filter["+name+"]", strings.Join(values, ","))
		}
	}
	for resourceType, fieldNames := range q.fields {
		if len(fieldNames) > 0 {
			v.Set("fields["+resourceType+"]", strings.Join(fieldNames, ","))
		}
	}
	if len(q.include) > 0 {
		v.Set("include", strings.Join(q.include, ","))
	}
	if len(q.sort) > 0 {
		v.Set("sort", strings.Join(q.sort, ","))
	}
	if q.hasLimit {
		v.Set("limit", strconv.Itoa(q.limit))
	}
	if q.cursor != "" {
		v.Set("cursor", q.cursor)
	}
	for key, values := range q.extra {
		for _, val := range values {
			v.Add(key, val)
		}
	}
	// url.Values.Encode already sorts keys alphabetically, which is what
	// we want for deterministic output.
	return v.Encode()
}

// sortedKeys is a small helper exported for tests that want to iterate
// filter/field keys deterministically.
func (q *Query) sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (q *Query) ensure() {
	if q.filters == nil {
		q.filters = make(map[string][]string)
	}
	if q.fields == nil {
		q.fields = make(map[string][]string)
	}
	if q.extra == nil {
		q.extra = make(map[string][]string)
	}
}

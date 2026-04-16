// Package parser converts Apple's App Store Connect OpenAPI JSON into
// the generator's intermediate representation (ir.Document).
//
// Parse is the sole public entry point. It is split into private
// stages (decodeRaw, extractResources, buildOperations) so each
// stage can be unit-tested against the minimal fixture; the real
// Apple spec is exercised end-to-end in Task 11.
package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/ir"
	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/naming"
	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/skip"
)

// Option tunes Parse behavior. Options compose via successive calls.
type Option func(*options)

type options struct {
	skip *skip.Set
}

// WithSkipSet filters out resources whose API name appears in set.
// A nil set means "skip nothing", matching the default.
func WithSkipSet(set *skip.Set) Option {
	return func(o *options) { o.skip = set }
}

// Parse turns OpenAPI JSON bytes into an *ir.Document. See
// package-level docs for a description of each stage.
func Parse(data []byte, opts ...Option) (*ir.Document, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}
	raw, err := decodeRaw(data)
	if err != nil {
		return nil, err
	}
	resources := extractResources(raw)
	if cfg.skip != nil {
		resources = filterSkipped(resources, cfg.skip)
	}
	return &ir.Document{Resources: resources}, nil
}

func filterSkipped(rs []ir.Resource, set *skip.Set) []ir.Resource {
	// Allocate a fresh slice rather than reusing rs[:0] — the latter
	// silently mutates the caller's backing array if they ever keep
	// a reference, which is a well-known footgun in code review.
	out := make([]ir.Resource, 0, len(rs))
	for _, r := range rs {
		if set.Contains(r.APIName) {
			continue
		}
		out = append(out, r)
	}
	return out
}

// rawSpec is a minimal-surface OpenAPI 3 document — only the fields
// we currently read. Unknown keys are ignored.
type rawSpec struct {
	OpenAPI string                     `json:"openapi"`
	Info    rawInfo                    `json:"info"`
	Paths   map[string]json.RawMessage `json:"paths"`
}

type rawInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

// rawPathItem is an OpenAPI Path Item Object. We model each HTTP verb
// as an optional rawOperation.
type rawPathItem struct {
	Get    *rawOperation `json:"get,omitempty"`
	Put    *rawOperation `json:"put,omitempty"`
	Post   *rawOperation `json:"post,omitempty"`
	Delete *rawOperation `json:"delete,omitempty"`
	Patch  *rawOperation `json:"patch,omitempty"`
}

type rawOperation struct {
	OperationID  string          `json:"operationId"`
	Summary      string          `json:"summary"`
	Tags         []string        `json:"tags"`
	Deprecated   bool            `json:"deprecated"`
	ExternalDocs rawExternalDocs `json:"externalDocs"`
}

type rawExternalDocs struct {
	URL string `json:"url"`
}

func decodeRaw(data []byte) (*rawSpec, error) {
	var r rawSpec
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parser: unmarshal spec: %w", err)
	}
	if r.OpenAPI == "" {
		return nil, fmt.Errorf("parser: missing top-level 'openapi' version")
	}
	if len(r.Paths) == 0 {
		return nil, fmt.Errorf("parser: no paths found in spec")
	}
	return &r, nil
}

// extractResources groups paths by their first "/v1/<resource>"
// segment. Each group becomes one ir.Resource. Operations within the
// group are attached with canonical names derived from HTTP verb +
// path shape (List/Get/Create/Update/Replace/Delete) plus a
// verb+PascalCase(tail) fallback for sub-resource paths.
func extractResources(raw *rawSpec) []ir.Resource {
	groups := make(map[string][]pathOp)
	for pathTmpl, rawItem := range raw.Paths {
		apiName := resourceFromPath(pathTmpl)
		if apiName == "" {
			continue
		}
		var item rawPathItem
		if err := json.Unmarshal(rawItem, &item); err != nil {
			// Generator-grade input: log the malformed path so Task 11's
			// smoke run produces a visible signal, but keep going so a
			// single bad item doesn't tank the whole parse.
			log.Printf("parser: skipping malformed path item %q: %v", pathTmpl, err)
			continue
		}
		for verb, op := range verbOps(&item) {
			if op == nil {
				continue
			}
			groups[apiName] = append(groups[apiName], pathOp{
				Path: pathTmpl,
				Verb: verb,
				Op:   op,
			})
		}
	}

	names := make([]string, 0, len(groups))
	for k := range groups {
		names = append(names, k)
	}
	sort.Strings(names)

	out := make([]ir.Resource, 0, len(names))
	for _, apiName := range names {
		ops := buildOperations(apiName, groups[apiName])
		out = append(out, ir.Resource{
			Name:       naming.PascalCase(apiName),
			APIName:    apiName,
			Operations: ops,
		})
	}
	return out
}

// pathOp is a single HTTP verb + path template + raw operation body.
type pathOp struct {
	Path string
	Verb string
	Op   *rawOperation
}

func verbOps(item *rawPathItem) map[string]*rawOperation {
	return map[string]*rawOperation{
		"GET":    item.Get,
		"PUT":    item.Put,
		"POST":   item.Post,
		"DELETE": item.Delete,
		"PATCH":  item.Patch,
	}
}

// resourceFromPath returns the resource name from a "/v1/foo/..."
// template, or "" for non-v1 paths. It ignores path parameters and
// sub-paths so "/v1/apps/{id}/relationships/appInfos" still maps to
// "apps" — that's sufficient for Phase 6.1 counting; sub-resource
// operations get distinguished by name in buildOperations.
func resourceFromPath(p string) string {
	p = strings.TrimPrefix(p, "/")
	parts := strings.Split(p, "/")
	if len(parts) < 2 || parts[0] != "v1" {
		return ""
	}
	name := parts[1]
	if strings.HasPrefix(name, "{") {
		return ""
	}
	return name
}

// buildOperations converts a grouped list of path+verb+raw-op into
// named ir.Operation entries. Naming rules are deliberately
// conservative in Phase 6.1; Phase 6.3 will refine them once we see
// how Apple names everything in practice.
func buildOperations(apiName string, ops []pathOp) []ir.Operation {
	sort.SliceStable(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Verb < ops[j].Verb
	})

	out := make([]ir.Operation, 0, len(ops))
	for _, po := range ops {
		name := operationName(apiName, po)
		out = append(out, ir.Operation{
			Name:         name,
			HTTPMethod:   po.Verb,
			PathTemplate: po.Path,
			Summary:      po.Op.Summary,
			Deprecated:   po.Op.Deprecated,
			DocURL:       po.Op.ExternalDocs.URL,
		})
	}
	return out
}

// operationName chooses a Go method name from the verb + path.
//
//	GET  /v1/apps                 -> "List"
//	GET  /v1/apps/{id}            -> "Get"
//	POST /v1/apps                 -> "Create"
//	PATCH /v1/apps/{id}           -> "Update"
//	DELETE /v1/apps/{id}          -> "Delete"
//	any sub-path                  -> verb + PascalCase(tail) fallback
func operationName(apiName string, po pathOp) string {
	base := "/v1/" + apiName
	switch {
	case po.Path == base && po.Verb == "GET":
		return "List"
	case po.Path == base && po.Verb == "POST":
		return "Create"
	case po.Path == base+"/{id}" && po.Verb == "GET":
		return "Get"
	case po.Path == base+"/{id}" && po.Verb == "PATCH":
		return "Update"
	case po.Path == base+"/{id}" && po.Verb == "PUT":
		return "Replace"
	case po.Path == base+"/{id}" && po.Verb == "DELETE":
		return "Delete"
	}
	// Fallback: verb + PascalCase(tail) where tail excludes any path
	// parameter segments. {id} segments add no semantic value to a Go
	// method name — "modify the appInfos relationship for an app" should
	// be UpdateRelationshipsAppInfos, not UpdateIDRelationshipsAppInfos.
	tail := strings.TrimPrefix(po.Path, base)
	var words []string
	for _, seg := range strings.Split(tail, "/") {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			continue // skip path-parameter segments
		}
		words = append(words, seg)
	}
	var b strings.Builder
	b.WriteString(verbPrefix(po.Verb))
	for _, w := range words {
		b.WriteString(naming.PascalCase(w))
	}
	return b.String()
}

func verbPrefix(verb string) string {
	switch verb {
	case "GET":
		return "Get"
	case "POST":
		return "Create"
	case "PATCH":
		return "Update"
	case "PUT":
		return "Replace"
	case "DELETE":
		return "Delete"
	default:
		return naming.PascalCase(strings.ToLower(verb))
	}
}

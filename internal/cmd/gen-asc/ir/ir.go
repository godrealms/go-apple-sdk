// Package ir defines the intermediate representation consumed by the
// code generator. Parsers read OpenAPI JSON and produce a *Document;
// renderers (Phase 6.2+) walk the Document and emit Go source.
//
// All fields use JSON struct tags so a Document can be serialized
// via `gen-asc -dump-ir` for human inspection and test fixtures.
package ir

// Document is the top-level IR produced by parsing one OpenAPI spec.
type Document struct {
	Metadata  Metadata   `json:"metadata"`
	Resources []Resource `json:"resources"`
}

// Metadata captures provenance for the document.
type Metadata struct {
	SpecVersion string `json:"spec_version"`
	SpecSHA256  string `json:"spec_sha256"`
	SpecSource  string `json:"spec_source"`
	GeneratedAt string `json:"generated_at"` // RFC3339
}

// Resource is one JSON:API resource type exposed by Apple, e.g.
// "apps" or "analyticsReports". It aggregates all operations that
// share a resource path prefix, plus the attribute/relationship
// schema used to decode or encode its documents.
type Resource struct {
	Name       string         `json:"name"`       // Go: "AnalyticsReports"
	APIName    string         `json:"api_name"`   // OpenAPI: "analyticsReports"
	DocURL     string         `json:"doc_url,omitempty"`
	Operations []Operation    `json:"operations"`
	Attrs      []Field        `json:"attrs,omitempty"`
	Rels       []Relationship `json:"rels,omitempty"`
}

// Operation is one HTTP endpoint on a resource.
type Operation struct {
	Name         string  `json:"name"`       // Go method name, e.g. "List"
	HTTPMethod   string  `json:"http_method"`
	PathTemplate string  `json:"path_template"`
	PathParams   []Field `json:"path_params,omitempty"`
	QueryParams  []Field `json:"query_params,omitempty"`
	RequestBody  *Type   `json:"request_body,omitempty"`
	ResponseBody *Type   `json:"response_body,omitempty"`
	DocURL       string  `json:"doc_url,omitempty"`
	Deprecated   bool    `json:"deprecated,omitempty"`
	Summary      string  `json:"summary,omitempty"`
}

// Field is a named, typed piece of data — attribute, path/query param,
// or body field.
type Field struct {
	Name     string `json:"name"`      // Go: "BundleID"
	JSONName string `json:"json_name"` // JSON: "bundleId"
	GoType   string `json:"go_type"`   // e.g. "string", "*time.Time", "[]string"
	Required bool   `json:"required,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

// Type is a body schema (request or response) referenced by an
// Operation. Phase 6.1 may leave this nil for every operation — the
// parser populates it progressively as Phases 6.2/6.3 need richer
// information.
type Type struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields,omitempty"`
}

// Relationship is a JSON:API relationship entry on a Resource.
type Relationship struct {
	Name   string `json:"name"`       // Go: "App"
	Target string `json:"target"`     // OpenAPI resource name of the target
	ToMany bool   `json:"to_many,omitempty"`
}

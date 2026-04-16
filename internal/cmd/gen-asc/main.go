// Command gen-asc generates Go code for the App Store Connect API from
// Apple's vendored OpenAPI specification.
//
// This binary is internal tooling, not part of the public SDK surface:
// users of github.com/godrealms/go-apple-sdk never need to run it. It
// is invoked by maintainers via `go run ./internal/cmd/gen-asc` or the
// eventual `make gen` target.
//
// Phase 6.1 scope: parse the vendored spec into an intermediate
// representation (IR) and dump that IR as JSON, or print a human-
// readable resource table. No Go code is generated yet — that arrives
// in Phase 6.2 and later.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/ir"
	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/parser"
	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/skip"
	"github.com/godrealms/go-apple-sdk/internal/cmd/gen-asc/spec"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen-asc: ")

	var (
		mode      = flag.String("mode", "list", "one of: list, dump-ir")
		skipPath  = flag.String("skip", defaultSkipPath(), "path to skip.txt")
		applySkip = flag.Bool("apply-skip", true, "filter out hand-written resources")
		pretty    = flag.Bool("pretty", true, "pretty-print JSON output")
	)
	flag.Parse()

	doc, err := loadIR(*skipPath, *applySkip)
	if err != nil {
		log.Fatalf("load IR: %v", err)
	}

	switch *mode {
	case "list":
		printTable(doc)
	case "dump-ir":
		printJSON(doc, *pretty)
	default:
		log.Fatalf("unknown -mode %q (want: list, dump-ir)", *mode)
	}
}

func loadIR(skipPath string, applySkip bool) (*ir.Document, error) {
	var opts []parser.Option
	if applySkip {
		set, err := skip.Load(skipPath)
		if err != nil {
			return nil, fmt.Errorf("load skip: %w", err)
		}
		opts = append(opts, parser.WithSkipSet(set))
	}
	doc, err := parser.Parse(spec.File, opts...)
	if err != nil {
		return nil, err
	}
	doc.Metadata = ir.Metadata{
		SpecVersion: spec.SpecVersion,
		SpecSHA256:  spec.SpecSHA256,
		SpecSource:  spec.SpecSource,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	return doc, nil
}

func printJSON(doc *ir.Document, pretty bool) {
	enc := json.NewEncoder(os.Stdout)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(doc); err != nil {
		log.Fatalf("encode: %v", err)
	}
}

func printTable(doc *ir.Document) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "RESOURCE\tGO NAME\tOPERATIONS\n")
	// Deterministic output.
	sort.SliceStable(doc.Resources, func(i, j int) bool {
		return doc.Resources[i].APIName < doc.Resources[j].APIName
	})
	for _, r := range doc.Resources {
		fmt.Fprintf(tw, "%s\t%s\t%d\n", r.APIName, r.Name, len(r.Operations))
	}
	fmt.Fprintf(tw, "\nTotal: %d resources (skip applied).\n", len(doc.Resources))
}

// defaultSkipPath resolves skip.txt relative to this source file so
// `go run ./internal/cmd/gen-asc` works from the repo root.
func defaultSkipPath() string {
	return filepath.Join("internal", "cmd", "gen-asc", "skip", "skip.txt")
}

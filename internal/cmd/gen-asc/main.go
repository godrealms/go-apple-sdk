// Command gen-asc generates Go code for the App Store Connect API from
// Apple's vendored OpenAPI specification.
//
// This binary is internal tooling, not part of the public SDK surface:
// users of github.com/godrealms/go-apple-sdk never need to run it. It
// is invoked by maintainers via `go run ./internal/cmd/gen-asc` or the
// eventual `make gen` target.
//
// Phase 6.1 scope: parse the vendored spec into an intermediate
// representation (IR) and dump that IR as JSON. No Go code is generated
// yet — that arrives in Phase 6.2 and later. Task 10 of the Phase 6.1
// plan replaces the stub main below with the real CLI implementation.
package main

func main() {}

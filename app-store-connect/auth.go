package AppStoreConnect

import "net/http"

// Authorizer attaches authentication material to an outgoing request.
// Implementations are expected to set the "Authorization" header on req,
// typically to "Bearer <signed JWT>".
//
// Apple's JWT tokens have a short lifetime (20 minutes maximum); a typical
// implementation signs fresh JWTs on every call, or caches them for a few
// minutes and re-signs before expiry. The SDK calls Authorize for every
// request and does not cache.
type Authorizer interface {
	Authorize(req *http.Request) error
}

// AuthorizerFunc adapts a plain function to the [Authorizer] interface.
type AuthorizerFunc func(req *http.Request) error

// Authorize implements [Authorizer].
func (f AuthorizerFunc) Authorize(req *http.Request) error { return f(req) }

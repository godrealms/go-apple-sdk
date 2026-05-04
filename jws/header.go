package jws

// Alg is a JWS "alg" header value. The SDK only accepts "ES256".
// Defined as an alias so callers can write `if h.Alg == "ES256"`
// without a typed conversion.
type Alg = string

// Header is the protected JWS header — only the fields the SDK
// cares about. Unknown fields are tolerated by encoding/json.
type Header struct {
	Alg Alg `json:"alg"`
	X5c X5c `json:"x5c"`
}

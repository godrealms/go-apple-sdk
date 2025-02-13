package types

// Alg The JSON Web Signature (JWS) header parameter that identifies the cryptographic algorithm used to secure the JWS.
type Alg string

func (a Alg) String() string {
	return string(a)
}

package types

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

// The JSON Web Signature (JWS) header parameter that identifies the cryptographic algorithm used to secure the JWS.
type alg string

// The JSON Web Signature (JWS) header parameter that contains the certificate chain that corresponds to the key used to digitally sign the JWS.
type x5c []string

func (c x5c) GetPublicKey() (*x509.Certificate, error) {
	if len(c) == 0 {
		return nil, fmt.Errorf("no certificates found in x5c")
	}
	// Decode the first certificate in the chain
	certBytes, err := base64.StdEncoding.DecodeString(c[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %w", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

type JWSDecodedHeader struct {
	// The algorithm used for signing the JSON Web Signature (JWS).
	Alg alg `json:"alg"`
	// The X.509 certificate chain that corresponds to the key that the App Store used to secure the JWS.
	X5c x5c `json:"x5c"`
}

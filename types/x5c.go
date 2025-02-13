package types

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
)

// X5c The JSON Web Signature (JWS) header parameter that contains the certificate chain that corresponds to the key used to digitally sign the JWS.
type X5c []string

func (x X5c) String() string {
	return strings.Join(x, ",")
}

func (x X5c) GetPublicKey() (*x509.Certificate, error) {
	if len(x) == 0 {
		return nil, fmt.Errorf("no certificates found in x5c")
	}
	// Decode the first certificate in the chain
	certBytes, err := base64.StdEncoding.DecodeString(x[0])
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

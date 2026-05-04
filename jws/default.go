package jws

import (
	"crypto/x509"
	_ "embed"
	"fmt"
	"sync"
)

//go:embed apple_root_ca_g3.pem
var appleRootCAG3PEM []byte

var (
	defaultOnce sync.Once
	defaultV    *Verifier
	defaultErr  error
)

// DefaultVerifier returns the process-wide Verifier configured
// with the embedded Apple Root CA G3 and DefaultRequiredOIDs.
// The first call parses the embedded PEM under sync.Once;
// subsequent calls return the cached singleton.
//
// If the embedded PEM fails to parse, DefaultVerifier panics.
// The PEM is a build-time asset — failing to parse means the
// binary is corrupt or someone replaced the file with garbage,
// and there is no sensible fallback.
func DefaultVerifier() *Verifier {
	defaultOnce.Do(func() {
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(appleRootCAG3PEM) {
			defaultErr = fmt.Errorf("jws: embedded Apple Root CA G3 PEM is invalid")
			return
		}
		defaultV = NewVerifier(WithRootCAs(pool))
	})
	if defaultErr != nil {
		panic(defaultErr)
	}
	return defaultV
}

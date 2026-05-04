# JWS Chain Validation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add full RFC 5280 certificate chain validation + mandatory Apple OID check to all JWS verification paths in the SDK (App Store Server Notifications V2 + JWSTransaction + JWSRenewalInfo). Fixes a CRITICAL security gap where the current code accepts any leaf cert from any CA.

**Architecture:** New top-level `jws/` package with a `*Verifier` that owns root CA pool + required OID list + clock. Generic `VerifyAndDecode[T any]` is the single verification entry point — used by `DefaultVerifier()` (embedded Apple Root CA G3, `sync.Once` lazy init) for the public API back-compat path, or with a custom Verifier via new `DecryptWith` / `DecodedPayloadWith` overloads. Existing methods keep their signatures, so callers get the security upgrade for free on `go get -u`. Three duplicated `parseSignedPayload` implementations and the orphaned `types/x5c.go` are deleted as part of consolidation.

**Tech Stack:** Go 1.23.1 stdlib only (`crypto/x509`, `crypto/ecdsa`, `crypto/sha256`, `encoding/asn1`, `encoding/base64`, `encoding/json`, `embed`, `sync`). No new third-party dependencies.

**Spec:** [`docs/superpowers/specs/2026-05-05-jws-chain-validation-design.md`](../specs/2026-05-05-jws-chain-validation-design.md) — read this before starting.

---

## Pre-flight: Glossary

- **JWS** (RFC 7515): JSON Web Signature in compact form `headerB64.payloadB64.signatureB64`. Each part is base64url-encoded (no padding).
- **x5c**: JWS header field; an array of base64-**std**-encoded DER certificates. `x5c[0]` is the leaf.
- **Apple Root CA G3**: ECC P-384 root cert from 2014, valid until 2039. Apple distributes it at https://www.apple.com/certificateauthority/ . Signs all StoreKit-related JWS payloads via an intermediate cert (Apple Worldwide Developer Relations CA).
- **OID**: ASN.1 Object Identifier. Apple's receipt-signing leaf carries OID `1.2.840.113635.100.6.11.1` in its X.509 extensions list. The plan also probes `1.2.840.113635.100.6.29` (Server Notifications V2) — see Task 0.
- **ES256**: ECDSA with SHA-256 over P-256 curve. JWS represents the signature as IEEE P1363 raw bytes: `r ‖ s` each 32 bytes, total 64.

## Pre-flight: Conventions

- All file paths in this plan are repo-relative (`jws/x5c.go` means `/Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk/jws/x5c.go`).
- Every code-changing step shows the actual code or diff to apply.
- Every test step shows the exact `go test` invocation and the expected pass/fail signal.
- Tasks are independently committable. Run `go vet ./...` and `go test -race ./...` before each commit; if either fails, fix before committing.
- Module path: `github.com/godrealms/go-apple-sdk`.

---

## Task 0: Verify OID `1.2.840.113635.100.6.29` (BLOCKING — no code yet)

**Why blocking:** Spec §11 marks confirmation of this OID as a hard prerequisite. If the actual App Store Server Notifications V2 leaf cert does NOT carry `6.29`, our `DefaultRequiredOIDs` list will reject every legitimate notification.

**Files:**
- Modify: `docs/superpowers/specs/2026-05-05-jws-chain-validation-design.md` (only if OID list needs revision)
- Create: `docs/superpowers/notes/2026-05-05-oid-confirmation.md` (records the verification result either way)

- [ ] **Step 1: Obtain a real sandbox V2 notification's `signedPayload`**

  Easiest path: run the existing example, which uses Apple's RequestTestNotification API to ask Apple to send a test V2 notification.

  ```bash
  # Requires Apple Issuer ID + Key ID + .p8 private key configured in env vars
  # If you don't have credentials, ask the project owner for a captured sample and skip ahead to Step 3.
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk
  go run examples/app-store-server/notifications-testing/main.go > /tmp/test_notification.json
  ```

  Expected: a JSON document with a `signedPayload` field (the JWS string). If the example doesn't print the signedPayload, modify it temporarily to do so.

- [ ] **Step 2: Extract leaf certificate from the JWS header**

  ```bash
  # Pull JWS string out of the captured response. Adjust the jq path if needed.
  JWS=$(jq -r '.signedPayload' /tmp/test_notification.json)
  HEADER_B64=$(echo "$JWS" | cut -d'.' -f1)
  # base64url → base64 std (add padding, swap chars)
  HEADER_JSON=$(echo "$HEADER_B64" | tr '_-' '/+' | awk '{l=length($0); pad=4-(l%4); if(pad<4) for(i=0;i<pad;i++) $0=$0"="; print}' | base64 -d)
  echo "$HEADER_JSON" | jq .
  echo "$HEADER_JSON" | jq -r '.x5c[0]' | base64 -d > /tmp/leaf.der
  ```

  Expected: `/tmp/leaf.der` is ~1KB of binary DER.

- [ ] **Step 3: Dump the leaf cert and search for Apple OIDs**

  ```bash
  openssl x509 -inform DER -in /tmp/leaf.der -text -noout | \
      grep -E "1\.2\.840\.113635\.100\.6"
  ```

  Expected output looks something like:
  ```
              1.2.840.113635.100.6.11.1
              1.2.840.113635.100.6.29
  ```
  The first OID (`6.11.1`) is the well-documented Apple Receipt Signing OID — it should be present. The question is whether `6.29` (or any other `6.x`) is also present.

- [ ] **Step 4: Record the finding**

  Create `docs/superpowers/notes/2026-05-05-oid-confirmation.md`:

  ```markdown
  # OID confirmation for jws/ DefaultRequiredOIDs

  **Date:** 2026-05-05
  **Sample source:** Sandbox V2 notification captured via RequestTestNotification

  ## Observed OIDs in leaf cert

  - `1.2.840.113635.100.6.11.1` — Apple Receipt Signing — PRESENT
  - `1.2.840.113635.100.6.29`   — App Store Server Notifications V2 — [PRESENT / NOT PRESENT]
  - (any others starting with 1.2.840.113635.100.6.X — list them)

  ## Decision for jws.DefaultRequiredOIDs

  - [ ] Both `6.11.1` and `6.29` go in the default list (if both present)
  - [ ] Only `6.11.1` (if `6.29` not present); revisit if Apple ever rejects payloads carrying only the receipt OID
  - [ ] Other (explain)
  ```

  Fill in the brackets based on the openssl output.

- [ ] **Step 5: If `6.29` is NOT present, edit the spec**

  Open `docs/superpowers/specs/2026-05-05-jws-chain-validation-design.md` and replace §4.2's `DefaultRequiredOIDs` block to drop `OIDAppleNotificationSigning` from the default list. Subsequent tasks must use whatever Task 0 records as the canonical default OID list.

- [ ] **Step 6: Commit Task 0 result**

  ```bash
  git add docs/superpowers/notes/2026-05-05-oid-confirmation.md
  # Also add the spec if you edited it.
  git commit -m "docs(jws): confirm DefaultRequiredOIDs from sandbox notification

  Verified leaf cert OIDs by openssl-dumping a real sandbox V2
  notification. See note for the canonical default OID list used
  by jws.DefaultVerifier."
  ```

---

## Task 1: Initialize `jws/` package skeleton

**Files:**
- Create: `jws/doc.go`

- [ ] **Step 1: Create the package directory + doc.go**

  ```bash
  mkdir -p jws/internal/testchain
  ```

  Write `jws/doc.go`:
  ```go
  // Package jws verifies the JSON Web Signature payloads Apple sends
  // for App Store Server Notifications V2, signed transactions, and
  // signed renewal info.
  //
  // All verification goes through a *Verifier. Use DefaultVerifier for
  // the standard configuration (Apple Root CA G3 embedded into the
  // SDK, mandatory Apple receipt-signing OID check), or build a
  // custom one with NewVerifier + options for testing or future
  // root-cert rotation:
  //
  //   v := jws.NewVerifier(
  //       jws.WithRootCAs(myPool),
  //       jws.WithRequiredOIDs(myOID),
  //   )
  //   payload, err := jws.VerifyAndDecode[Transaction](v, raw)
  //
  // All failures return *VerificationError; switch on
  // VerificationError.Reason to distinguish chain failure from OID
  // mismatch from signature mismatch from malformed input.
  //
  // The package targets Apple's documented JWS profile (ES256 only,
  // x5c chain present, leaf carries Apple OID). It is intentionally
  // not a general-purpose JWS library.
  package jws
  ```

- [ ] **Step 2: Verify package compiles**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk
  go build ./jws/
  ```

  Expected: no output, exit 0.

- [ ] **Step 3: Commit**

  ```bash
  git add jws/doc.go
  git commit -m "feat(jws): bootstrap jws/ package with doc.go"
  ```

---

## Task 2: Implement `internal/testchain` helper

This helper lives under `internal/` so it is unimportable by callers but accessible to all tests in the `jws` package and in `types/`. It generates a fresh in-memory ECDSA chain (root P-384 → intermediate P-256 → leaf P-256) and signs JWS strings with the leaf key.

**Files:**
- Create: `jws/internal/testchain/testchain.go`
- Create: `jws/internal/testchain/testchain_test.go`

- [ ] **Step 1: Write failing test for `New` + `SignJWS` round-trip**

  Write `jws/internal/testchain/testchain_test.go`:
  ```go
  package testchain

  import (
      "crypto/ecdsa"
      "crypto/sha256"
      "encoding/asn1"
      "encoding/base64"
      "encoding/json"
      "math/big"
      "strings"
      "testing"
      "time"
  )

  func TestNew_BuildsValidChain(t *testing.T) {
      tc := New(t)
      if tc.Root == nil || tc.Intermediate == nil || tc.Leaf == nil {
          t.Fatalf("expected all three certs populated")
      }
      // Intermediate is signed by Root.
      if err := tc.Intermediate.CheckSignatureFrom(tc.Root); err != nil {
          t.Fatalf("intermediate not signed by root: %v", err)
      }
      // Leaf is signed by Intermediate.
      if err := tc.Leaf.CheckSignatureFrom(tc.Intermediate); err != nil {
          t.Fatalf("leaf not signed by intermediate: %v", err)
      }
      if tc.RootPool == nil {
          t.Fatalf("expected RootPool populated")
      }
  }

  func TestSignJWS_RoundTripVerifies(t *testing.T) {
      tc := New(t)
      type payload struct {
          Foo string `json:"foo"`
      }
      raw := tc.SignJWS(t, payload{Foo: "bar"})
      parts := strings.Split(raw, ".")
      if len(parts) != 3 {
          t.Fatalf("expected 3 JWS segments, got %d", len(parts))
      }
      // Verify signature externally to confirm SignJWS produced
      // valid IEEE P1363 ES256.
      headerJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
      var hdr struct {
          Alg string   `json:"alg"`
          X5c []string `json:"x5c"`
      }
      if err := json.Unmarshal(headerJSON, &hdr); err != nil {
          t.Fatalf("decode header: %v", err)
      }
      if hdr.Alg != "ES256" {
          t.Fatalf("expected alg ES256, got %q", hdr.Alg)
      }
      sig, _ := base64.RawURLEncoding.DecodeString(parts[2])
      if len(sig) != 64 {
          t.Fatalf("expected 64-byte raw signature, got %d", len(sig))
      }
      hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
      r := new(big.Int).SetBytes(sig[:32])
      s := new(big.Int).SetBytes(sig[32:])
      pub := tc.Leaf.PublicKey.(*ecdsa.PublicKey)
      if !ecdsa.Verify(pub, hash[:], r, s) {
          t.Fatalf("ecdsa verify failed on testchain output")
      }
  }

  func TestNew_WithLeafOIDs_AppliesExtensions(t *testing.T) {
      myOID := asn1.ObjectIdentifier{1, 2, 3, 4, 5}
      tc := New(t, WithLeafOIDs(myOID))
      var found bool
      for _, ext := range tc.Leaf.Extensions {
          if ext.Id.Equal(myOID) {
              found = true
              break
          }
      }
      if !found {
          t.Fatalf("expected custom OID in leaf extensions")
      }
  }

  func TestNew_WithLeafNotAfter_AppliesExpiration(t *testing.T) {
      past := time.Now().Add(-time.Hour)
      tc := New(t, WithLeafNotAfter(past))
      if !tc.Leaf.NotAfter.Equal(past) {
          t.Fatalf("expected leaf NotAfter %v, got %v", past, tc.Leaf.NotAfter)
      }
  }
  ```

- [ ] **Step 2: Run test, verify failure**

  ```bash
  go test ./jws/internal/testchain/...
  ```

  Expected: build error or `undefined: New`. Either way, fail.

- [ ] **Step 3: Implement `testchain.go`**

  Write `jws/internal/testchain/testchain.go`:
  ```go
  // Package testchain builds in-memory ECDSA certificate chains and
  // signs JWS payloads with them, for use by jws/ tests and by tests
  // in dependent packages (e.g. types/JWSTransaction migration tests).
  //
  // The package lives under internal/ on purpose: it is for SDK tests
  // only, never for downstream callers' production code.
  package testchain

  import (
      "crypto/ecdsa"
      "crypto/elliptic"
      "crypto/rand"
      "crypto/sha256"
      "crypto/x509"
      "crypto/x509/pkix"
      "encoding/asn1"
      "encoding/base64"
      "encoding/json"
      "math/big"
      "testing"
      "time"
  )

  // Chain is the output of New. It carries enough material to either
  // (a) sign a JWS via SignJWS, or (b) hand its RootPool to a Verifier
  // under test.
  type Chain struct {
      Root, Intermediate, Leaf *x509.Certificate
      LeafKey                  *ecdsa.PrivateKey
      RootPool                 *x509.CertPool
  }

  // Opt customises the leaf cert built by New.
  type Opt func(*config)

  type config struct {
      leafOIDs      []asn1.ObjectIdentifier
      leafNotBefore time.Time
      leafNotAfter  time.Time
  }

  // WithLeafOIDs adds the given OIDs as custom X.509 extensions on the
  // leaf cert. Each is encoded with an empty ASN.1 NULL value, which
  // matches how Apple writes its receipt-signing OID extension.
  func WithLeafOIDs(oids ...asn1.ObjectIdentifier) Opt {
      return func(c *config) { c.leafOIDs = oids }
  }

  // WithLeafNotBefore overrides the leaf cert's NotBefore.
  func WithLeafNotBefore(t time.Time) Opt {
      return func(c *config) { c.leafNotBefore = t }
  }

  // WithLeafNotAfter overrides the leaf cert's NotAfter.
  func WithLeafNotAfter(t time.Time) Opt {
      return func(c *config) { c.leafNotAfter = t }
  }

  // appleReceiptSigningOID is the default OID stamped on the test
  // leaf — matches the production Apple receipt-signing OID so the
  // jws.DefaultVerifier accepts test-chain payloads without
  // additional configuration. Tests that want to exercise the OID
  // failure path should override with WithLeafOIDs.
  var appleReceiptSigningOID = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}

  // New builds a fresh root → intermediate → leaf chain. All keys
  // and certs live in process memory only; nothing touches disk.
  func New(t *testing.T, opts ...Opt) *Chain {
      t.Helper()
      cfg := &config{
          leafOIDs:      []asn1.ObjectIdentifier{appleReceiptSigningOID},
          leafNotBefore: time.Now().Add(-time.Hour),
          leafNotAfter:  time.Now().Add(24 * time.Hour),
      }
      for _, opt := range opts {
          opt(cfg)
      }

      // Root: self-signed P-384 (mirrors Apple Root CA G3 curve).
      rootKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
      mustNoError(t, err, "generate root key")
      rootTpl := &x509.Certificate{
          SerialNumber:          big.NewInt(1),
          Subject:               pkix.Name{CommonName: "test root"},
          NotBefore:             time.Now().Add(-time.Hour),
          NotAfter:              time.Now().Add(365 * 24 * time.Hour),
          IsCA:                  true,
          KeyUsage:              x509.KeyUsageCertSign,
          BasicConstraintsValid: true,
      }
      rootDER, err := x509.CreateCertificate(rand.Reader, rootTpl, rootTpl, &rootKey.PublicKey, rootKey)
      mustNoError(t, err, "create root cert")
      root, err := x509.ParseCertificate(rootDER)
      mustNoError(t, err, "parse root cert")

      // Intermediate: P-256, signed by root.
      intKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
      mustNoError(t, err, "generate intermediate key")
      intTpl := &x509.Certificate{
          SerialNumber:          big.NewInt(2),
          Subject:               pkix.Name{CommonName: "test intermediate"},
          NotBefore:             time.Now().Add(-time.Hour),
          NotAfter:              time.Now().Add(365 * 24 * time.Hour),
          IsCA:                  true,
          KeyUsage:              x509.KeyUsageCertSign,
          BasicConstraintsValid: true,
      }
      intDER, err := x509.CreateCertificate(rand.Reader, intTpl, root, &intKey.PublicKey, rootKey)
      mustNoError(t, err, "create intermediate cert")
      intermediate, err := x509.ParseCertificate(intDER)
      mustNoError(t, err, "parse intermediate cert")

      // Leaf: P-256, signed by intermediate.
      leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
      mustNoError(t, err, "generate leaf key")
      var extras []pkix.Extension
      for _, oid := range cfg.leafOIDs {
          extras = append(extras, pkix.Extension{
              Id:       oid,
              Critical: false,
              // ASN.1 NULL value (0x05 0x00) — minimal valid encoding.
              Value: []byte{0x05, 0x00},
          })
      }
      leafTpl := &x509.Certificate{
          SerialNumber:    big.NewInt(3),
          Subject:         pkix.Name{CommonName: "test leaf"},
          NotBefore:       cfg.leafNotBefore,
          NotAfter:        cfg.leafNotAfter,
          KeyUsage:        x509.KeyUsageDigitalSignature,
          ExtraExtensions: extras,
      }
      leafDER, err := x509.CreateCertificate(rand.Reader, leafTpl, intermediate, &leafKey.PublicKey, intKey)
      mustNoError(t, err, "create leaf cert")
      leaf, err := x509.ParseCertificate(leafDER)
      mustNoError(t, err, "parse leaf cert")

      pool := x509.NewCertPool()
      pool.AddCert(root)

      return &Chain{
          Root:         root,
          Intermediate: intermediate,
          Leaf:         leaf,
          LeafKey:      leafKey,
          RootPool:     pool,
      }
  }

  // SignJWS builds a JWS string (header.payload.signature) using the
  // leaf key. payload is JSON-marshalled; header carries alg=ES256 and
  // x5c=[leaf, intermediate, root] (each base64-std-encoded DER).
  func (c *Chain) SignJWS(t *testing.T, payload any) string {
      t.Helper()
      header := struct {
          Alg string   `json:"alg"`
          X5c []string `json:"x5c"`
      }{
          Alg: "ES256",
          X5c: []string{
              base64.StdEncoding.EncodeToString(c.Leaf.Raw),
              base64.StdEncoding.EncodeToString(c.Intermediate.Raw),
              base64.StdEncoding.EncodeToString(c.Root.Raw),
          },
      }
      headerJSON, err := json.Marshal(header)
      mustNoError(t, err, "marshal header")
      headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

      payloadJSON, err := json.Marshal(payload)
      mustNoError(t, err, "marshal payload")
      payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

      signingInput := headerB64 + "." + payloadB64
      hash := sha256.Sum256([]byte(signingInput))
      r, s, err := ecdsa.Sign(rand.Reader, c.LeafKey, hash[:])
      mustNoError(t, err, "ecdsa sign")
      sig := make([]byte, 64)
      r.FillBytes(sig[:32])
      s.FillBytes(sig[32:])
      sigB64 := base64.RawURLEncoding.EncodeToString(sig)

      return signingInput + "." + sigB64
  }

  func mustNoError(t *testing.T, err error, what string) {
      t.Helper()
      if err != nil {
          t.Fatalf("%s: %v", what, err)
      }
  }
  ```

- [ ] **Step 4: Run tests, verify pass**

  ```bash
  go test -race ./jws/internal/testchain/...
  ```

  Expected: `ok  github.com/godrealms/go-apple-sdk/jws/internal/testchain`.

- [ ] **Step 5: Commit**

  ```bash
  git add jws/internal/testchain/
  git commit -m "feat(jws/testchain): in-memory ECDSA chain + JWS signer for tests"
  ```

---

## Task 3: Implement `VerificationError` + `ReasonCode`

**Files:**
- Create: `jws/errors.go`
- Create: `jws/errors_test.go`

- [ ] **Step 1: Write failing tests**

  Write `jws/errors_test.go`:
  ```go
  package jws

  import (
      "errors"
      "fmt"
      "testing"
  )

  func TestVerificationError_Error(t *testing.T) {
      cases := []struct {
          name string
          err  *VerificationError
          want string
      }{
          {"with cause",
              &VerificationError{Reason: ReasonChain, Cause: fmt.Errorf("boom")},
              "jws: chain: boom"},
          {"without cause",
              &VerificationError{Reason: ReasonOID},
              "jws: oid"},
      }
      for _, c := range cases {
          t.Run(c.name, func(t *testing.T) {
              if got := c.err.Error(); got != c.want {
                  t.Fatalf("Error() = %q, want %q", got, c.want)
              }
          })
      }
  }

  func TestVerificationError_Unwrap(t *testing.T) {
      sentinel := fmt.Errorf("inner")
      ve := &VerificationError{Reason: ReasonSignature, Cause: sentinel}
      if !errors.Is(ve, sentinel) {
          t.Fatalf("errors.Is should match wrapped cause")
      }
  }

  func TestReasonCode_String(t *testing.T) {
      cases := map[ReasonCode]string{
          ReasonStructure: "structure",
          ReasonChain:     "chain",
          ReasonOID:       "oid",
          ReasonExpired:   "expired",
          ReasonSignature: "signature",
          ReasonCode(99):  "unknown",
      }
      for code, want := range cases {
          if got := code.String(); got != want {
              t.Fatalf("ReasonCode(%d).String() = %q, want %q", code, got, want)
          }
      }
  }
  ```

- [ ] **Step 2: Run tests, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: build errors — `undefined: VerificationError`, `undefined: ReasonCode`, etc.

- [ ] **Step 3: Implement `errors.go`**

  Write `jws/errors.go`:
  ```go
  package jws

  import "fmt"

  // ReasonCode is the kind of failure VerifyAndDecode encountered.
  // It is exposed so callers can branch on the failure category for
  // alerting and metrics without parsing error strings.
  type ReasonCode int

  const (
      // ReasonStructure means the JWS was malformed: wrong number of
      // segments, bad base64, undecodable header or payload, or the
      // header carried an algorithm other than ES256.
      ReasonStructure ReasonCode = iota + 1
      // ReasonChain means the leaf certificate did not chain back to
      // a trusted root via the intermediates supplied in x5c.
      ReasonChain
      // ReasonOID means the leaf certificate did not carry any of
      // the OIDs the Verifier required.
      ReasonOID
      // ReasonExpired means at least one certificate in the chain was
      // outside its validity window at verification time.
      ReasonExpired
      // ReasonSignature means the leaf cert public key did not verify
      // the JWS signature, or the signature bytes were malformed
      // (wrong length, non-ECDSA leaf key).
      ReasonSignature
  )

  // String returns the lowercase reason name used in error messages.
  func (r ReasonCode) String() string {
      switch r {
      case ReasonStructure:
          return "structure"
      case ReasonChain:
          return "chain"
      case ReasonOID:
          return "oid"
      case ReasonExpired:
          return "expired"
      case ReasonSignature:
          return "signature"
      default:
          return "unknown"
      }
  }

  // VerificationError is the only error type returned from this
  // package. Inspect VerificationError.Reason to branch on failure
  // class; use errors.Is/As against Cause for finer details (the
  // wrapped error is typically an *x509 error or a JSON decode
  // error).
  type VerificationError struct {
      Reason ReasonCode
      Cause  error // optional; nil when the failure has no underlying error
  }

  // Error implements error. Format: "jws: <reason>[: <cause>]".
  func (e *VerificationError) Error() string {
      if e.Cause == nil {
          return fmt.Sprintf("jws: %s", e.Reason)
      }
      return fmt.Sprintf("jws: %s: %v", e.Reason, e.Cause)
  }

  // Unwrap returns the wrapped cause so errors.Is and errors.As walk
  // through to it.
  func (e *VerificationError) Unwrap() error { return e.Cause }
  ```

- [ ] **Step 4: Run tests, verify pass**

  ```bash
  go test -race ./jws/
  ```

  Expected: `ok  github.com/godrealms/go-apple-sdk/jws`.

- [ ] **Step 5: Commit**

  ```bash
  git add jws/errors.go jws/errors_test.go
  git commit -m "feat(jws): VerificationError + ReasonCode"
  ```

---

## Task 4: Implement `X5c` + `Parse`

**Files:**
- Create: `jws/x5c.go`
- Create: `jws/x5c_test.go`

- [ ] **Step 1: Write failing tests**

  Write `jws/x5c_test.go`:
  ```go
  package jws

  import (
      "encoding/base64"
      "testing"

      "github.com/godrealms/go-apple-sdk/jws/internal/testchain"
  )

  func TestX5c_Parse_Success(t *testing.T) {
      tc := testchain.New(t)
      x := X5c{
          base64.StdEncoding.EncodeToString(tc.Leaf.Raw),
          base64.StdEncoding.EncodeToString(tc.Intermediate.Raw),
          base64.StdEncoding.EncodeToString(tc.Root.Raw),
      }
      certs, err := x.Parse()
      if err != nil {
          t.Fatalf("Parse: %v", err)
      }
      if len(certs) != 3 {
          t.Fatalf("expected 3 certs, got %d", len(certs))
      }
      if !certs[0].Equal(tc.Leaf) {
          t.Fatalf("certs[0] != leaf")
      }
  }

  func TestX5c_Parse_EmptyChain(t *testing.T) {
      var x X5c
      _, err := x.Parse()
      assertReason(t, err, ReasonStructure)
  }

  func TestX5c_Parse_BadBase64(t *testing.T) {
      x := X5c{"!!!!not-base64!!!!"}
      _, err := x.Parse()
      assertReason(t, err, ReasonStructure)
  }

  func TestX5c_Parse_BadDER(t *testing.T) {
      x := X5c{base64.StdEncoding.EncodeToString([]byte("not a cert"))}
      _, err := x.Parse()
      assertReason(t, err, ReasonStructure)
  }

  // assertReason is a shared helper. Defined here, used everywhere.
  func assertReason(t *testing.T, err error, want ReasonCode) {
      t.Helper()
      if err == nil {
          t.Fatalf("expected error with reason %s, got nil", want)
      }
      ve, ok := err.(*VerificationError)
      if !ok {
          t.Fatalf("expected *VerificationError, got %T: %v", err, err)
      }
      if ve.Reason != want {
          t.Fatalf("expected reason %s, got %s (cause=%v)", want, ve.Reason, ve.Cause)
      }
  }
  ```

- [ ] **Step 2: Run tests, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: build error — `undefined: X5c`.

- [ ] **Step 3: Implement `x5c.go`**

  Write `jws/x5c.go`:
  ```go
  package jws

  import (
      "crypto/x509"
      "encoding/base64"
      "fmt"
  )

  // X5c is the JWS x5c header field: an array of base64-std-encoded
  // DER certificates. By RFC 7515 convention, X5c[0] is the leaf
  // (signing) cert and the rest are intermediates in chain order.
  type X5c []string

  // Parse decodes each entry into a *x509.Certificate. Returns a
  // *VerificationError with ReasonStructure on any failure, including
  // an empty chain.
  func (x X5c) Parse() ([]*x509.Certificate, error) {
      if len(x) == 0 {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("x5c: empty chain"),
          }
      }
      out := make([]*x509.Certificate, 0, len(x))
      for i, b64 := range x {
          der, err := base64.StdEncoding.DecodeString(b64)
          if err != nil {
              return nil, &VerificationError{
                  Reason: ReasonStructure,
                  Cause:  fmt.Errorf("x5c[%d]: base64: %w", i, err),
              }
          }
          cert, err := x509.ParseCertificate(der)
          if err != nil {
              return nil, &VerificationError{
                  Reason: ReasonStructure,
                  Cause:  fmt.Errorf("x5c[%d]: parse: %w", i, err),
              }
          }
          out = append(out, cert)
      }
      return out, nil
  }
  ```

- [ ] **Step 4: Run tests, verify pass**

  ```bash
  go test -race ./jws/
  ```

  Expected: PASS.

- [ ] **Step 5: Commit**

  ```bash
  git add jws/x5c.go jws/x5c_test.go
  git commit -m "feat(jws): X5c type with Parse()"
  ```

---

## Task 5: Implement `Header` type

**Files:**
- Create: `jws/header.go`
- Create: `jws/header_test.go`

- [ ] **Step 1: Write failing tests**

  Write `jws/header_test.go`:
  ```go
  package jws

  import (
      "encoding/json"
      "testing"
  )

  func TestHeader_Unmarshal(t *testing.T) {
      raw := `{"alg":"ES256","x5c":["abc","def"]}`
      var h Header
      if err := json.Unmarshal([]byte(raw), &h); err != nil {
          t.Fatalf("unmarshal: %v", err)
      }
      if h.Alg != "ES256" {
          t.Fatalf("Alg = %q, want ES256", h.Alg)
      }
      if len(h.X5c) != 2 || h.X5c[0] != "abc" {
          t.Fatalf("X5c = %v, want [abc def]", h.X5c)
      }
  }
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: `undefined: Header`.

- [ ] **Step 3: Implement `header.go`**

  Write `jws/header.go`:
  ```go
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
  ```

- [ ] **Step 4: Run, verify pass**

  ```bash
  go test -race ./jws/
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add jws/header.go jws/header_test.go
  git commit -m "feat(jws): Header struct + Alg alias"
  ```

---

## Task 6: OID constants + `matchOID` helper

**Files:**
- Create: `jws/oid.go`
- Create: `jws/oid_test.go`

> **NOTE:** `DefaultRequiredOIDs` below uses both `OIDAppleReceiptSigning` and `OIDAppleNotificationSigning`. If Task 0 found `6.29` is NOT carried by real V2 notification leaves, drop `OIDAppleNotificationSigning` from `DefaultRequiredOIDs` (keep the constant defined for future use).

- [ ] **Step 1: Write failing tests**

  Write `jws/oid_test.go`:
  ```go
  package jws

  import (
      "crypto/x509/pkix"
      "encoding/asn1"
      "testing"
  )

  func TestMatchOID_Found(t *testing.T) {
      cert := &fakeCert{exts: []pkix.Extension{
          {Id: OIDAppleReceiptSigning},
      }}
      if !matchOID(certExtensions(cert), []asn1.ObjectIdentifier{OIDAppleReceiptSigning}) {
          t.Fatalf("expected match")
      }
  }

  func TestMatchOID_NotFound(t *testing.T) {
      cert := &fakeCert{exts: []pkix.Extension{
          {Id: asn1.ObjectIdentifier{1, 2, 3}},
      }}
      if matchOID(certExtensions(cert), []asn1.ObjectIdentifier{OIDAppleReceiptSigning}) {
          t.Fatalf("expected no match")
      }
  }

  func TestMatchOID_AnyOfRequired(t *testing.T) {
      cert := &fakeCert{exts: []pkix.Extension{
          {Id: OIDAppleNotificationSigning},
      }}
      // Required list contains both — any one is enough.
      if !matchOID(certExtensions(cert), []asn1.ObjectIdentifier{
          OIDAppleReceiptSigning, OIDAppleNotificationSigning,
      }) {
          t.Fatalf("expected match (notification OID present)")
      }
  }

  // fakeCert is a minimal stand-in for *x509.Certificate so the OID
  // tests don't need real cert generation. Production code passes
  // certificate.Extensions directly.
  type fakeCert struct {
      exts []pkix.Extension
  }

  func certExtensions(c *fakeCert) []pkix.Extension { return c.exts }
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: `undefined: OIDAppleReceiptSigning`, `undefined: matchOID`.

- [ ] **Step 3: Implement `oid.go`**

  Write `jws/oid.go`:
  ```go
  package jws

  import (
      "crypto/x509/pkix"
      "encoding/asn1"
  )

  // OIDAppleReceiptSigning is the X.509 OID Apple stamps on every
  // receipt-signing leaf certificate. Documented since iOS 7.
  var OIDAppleReceiptSigning = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 11, 1}

  // OIDAppleNotificationSigning was probed in Task 0 against a real
  // sandbox V2 notification's leaf cert. See
  // docs/superpowers/notes/2026-05-05-oid-confirmation.md for the
  // verification result. Keep the constant defined regardless of the
  // outcome; only DefaultRequiredOIDs below is conditional.
  var OIDAppleNotificationSigning = asn1.ObjectIdentifier{1, 2, 840, 113635, 100, 6, 29}

  // DefaultRequiredOIDs lists the OIDs DefaultVerifier requires the
  // leaf cert to carry. A leaf passes if its Extensions list contains
  // ANY of these (logical OR), not all of them.
  //
  // If Task 0 found that real notification leaves do not carry
  // OIDAppleNotificationSigning, drop it from this slice.
  var DefaultRequiredOIDs = []asn1.ObjectIdentifier{
      OIDAppleReceiptSigning,
      OIDAppleNotificationSigning,
  }

  // matchOID returns true if any extension's Id appears in required.
  // matchOID is a free function (not a method on *x509.Certificate)
  // so unit tests can pass synthetic extension lists.
  func matchOID(extensions []pkix.Extension, required []asn1.ObjectIdentifier) bool {
      for _, ext := range extensions {
          for _, req := range required {
              if ext.Id.Equal(req) {
                  return true
              }
          }
      }
      return false
  }
  ```

- [ ] **Step 4: Run, verify pass**

  ```bash
  go test -race ./jws/
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add jws/oid.go jws/oid_test.go
  git commit -m "feat(jws): Apple receipt/notification OIDs + matchOID"
  ```

---

## Task 7: Chain validation function

**Files:**
- Create: `jws/chain.go`
- Create: `jws/chain_test.go`

- [ ] **Step 1: Write failing tests**

  Write `jws/chain_test.go`:
  ```go
  package jws

  import (
      "crypto/x509"
      "testing"
      "time"

      "github.com/godrealms/go-apple-sdk/jws/internal/testchain"
  )

  func TestVerifyChain_Success(t *testing.T) {
      tc := testchain.New(t)
      err := verifyChain(
          []*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
          tc.RootPool,
          time.Now(),
      )
      if err != nil {
          t.Fatalf("expected nil, got %v", err)
      }
  }

  func TestVerifyChain_WrongRoot(t *testing.T) {
      tc := testchain.New(t)
      otherRoots := x509.NewCertPool() // empty pool
      err := verifyChain(
          []*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
          otherRoots,
          time.Now(),
      )
      assertReason(t, err, ReasonChain)
  }

  func TestVerifyChain_LeafExpired(t *testing.T) {
      tc := testchain.New(t, testchain.WithLeafNotAfter(time.Now().Add(-time.Hour)))
      err := verifyChain(
          []*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
          tc.RootPool,
          time.Now(),
      )
      assertReason(t, err, ReasonExpired)
  }

  func TestVerifyChain_LeafNotYetValid(t *testing.T) {
      tc := testchain.New(t, testchain.WithLeafNotBefore(time.Now().Add(time.Hour)))
      err := verifyChain(
          []*x509.Certificate{tc.Leaf, tc.Intermediate, tc.Root},
          tc.RootPool,
          time.Now(),
      )
      assertReason(t, err, ReasonExpired)
  }
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: `undefined: verifyChain`.

- [ ] **Step 3: Implement `chain.go`**

  Write `jws/chain.go`:
  ```go
  package jws

  import (
      "crypto/x509"
      "errors"
      "time"
  )

  // verifyChain runs RFC 5280 path validation on a JWS x5c chain.
  //
  // chain[0] must be the leaf; chain[1:] are intermediates (in any
  // order — the pool deduplicates). roots is the trust anchor pool.
  // now is the verification clock; passing time.Now() is normal,
  // passing a fixed instant is for tests.
  //
  // Returns nil on success, or *VerificationError with one of:
  //   ReasonExpired — when the chain is rejected for time-window
  //                   reasons (NotBefore in the future, NotAfter in
  //                   the past)
  //   ReasonChain   — for any other path-validation failure (unknown
  //                   authority, name mismatch, key usage, etc.)
  func verifyChain(chain []*x509.Certificate, roots *x509.CertPool, now time.Time) error {
      if len(chain) == 0 {
          return &VerificationError{
              Reason: ReasonStructure,
              Cause:  errors.New("chain: empty"),
          }
      }
      intermediates := x509.NewCertPool()
      for _, c := range chain[1:] {
          intermediates.AddCert(c)
      }
      _, err := chain[0].Verify(x509.VerifyOptions{
          Roots:         roots,
          Intermediates: intermediates,
          CurrentTime:   now,
      })
      if err == nil {
          return nil
      }
      // Map x509.CertificateInvalidError{Expired,NotYetValid} to
      // ReasonExpired so callers can branch on it. All other x509
      // errors map to ReasonChain.
      var invErr x509.CertificateInvalidError
      if errors.As(err, &invErr) {
          if invErr.Reason == x509.Expired || invErr.Reason == x509.NotYetValid {
              return &VerificationError{Reason: ReasonExpired, Cause: err}
          }
      }
      return &VerificationError{Reason: ReasonChain, Cause: err}
  }
  ```

- [ ] **Step 4: Run, verify pass**

  ```bash
  go test -race ./jws/
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add jws/chain.go jws/chain_test.go
  git commit -m "feat(jws): chain validation with Expired vs Chain mapping"
  ```

---

## Task 8: Signature verification function

This is internal to `jws/`; tested indirectly through `VerifyAndDecode` in Task 10. We pull it out so the `decode.go` orchestrator stays readable.

**Files:**
- Modify: `jws/chain.go` (append `verifySignature` function — co-located because both are crypto primitives)

- [ ] **Step 1: Write failing test (in chain_test.go)**

  Append to `jws/chain_test.go`:
  ```go
  func TestVerifySignature_Success(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, map[string]string{"hello": "world"})
      parts := splitJWS(t, raw)
      err := verifySignature(tc.Leaf, parts.headerB64, parts.payloadB64, parts.sig)
      if err != nil {
          t.Fatalf("expected nil, got %v", err)
      }
  }

  func TestVerifySignature_TamperedPayload(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, map[string]string{"hello": "world"})
      parts := splitJWS(t, raw)
      // Flip one byte in payloadB64 → the hashed input changes →
      // signature no longer verifies.
      tampered := []byte(parts.payloadB64)
      tampered[0] ^= 1
      err := verifySignature(tc.Leaf, parts.headerB64, string(tampered), parts.sig)
      assertReason(t, err, ReasonSignature)
  }

  func TestVerifySignature_TruncatedSignature(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, map[string]string{"hello": "world"})
      parts := splitJWS(t, raw)
      err := verifySignature(tc.Leaf, parts.headerB64, parts.payloadB64, parts.sig[:60])
      assertReason(t, err, ReasonSignature)
  }

  func TestVerifySignature_NonECDSAKey(t *testing.T) {
      // Use a cert whose PublicKey is RSA. We can't easily build one
      // with testchain (it's ECDSA-only), so synthesize a minimal
      // *x509.Certificate with an *rsa.PublicKey.
      rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
      cert := &x509.Certificate{PublicKey: &rsaKey.PublicKey}
      err := verifySignature(cert, "header", "payload", make([]byte, 64))
      assertReason(t, err, ReasonSignature)
  }

  // splitJWS is a test helper used by signature tests and by
  // decode_test.go.
  type jwsParts struct {
      headerB64, payloadB64 string
      sig                   []byte
  }

  func splitJWS(t *testing.T, raw string) jwsParts {
      t.Helper()
      bits := strings.SplitN(raw, ".", 3)
      if len(bits) != 3 {
          t.Fatalf("expected 3 segments, got %d", len(bits))
      }
      sig, err := base64.RawURLEncoding.DecodeString(bits[2])
      if err != nil {
          t.Fatalf("decode sig: %v", err)
      }
      return jwsParts{headerB64: bits[0], payloadB64: bits[1], sig: sig}
  }
  ```

  Also add the missing imports at the top of the file:
  ```go
  import (
      "crypto/rand"
      "crypto/rsa"
      "encoding/base64"
      "strings"
      // ...existing imports...
  )
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: `undefined: verifySignature`.

- [ ] **Step 3: Append `verifySignature` to `chain.go`**

  Add to `jws/chain.go`:
  ```go
  import (
      "crypto/ecdsa"
      "crypto/sha256"
      "math/big"
      // ...existing imports...
  )

  // verifySignature verifies a JWS ES256 signature using the leaf
  // cert's public key.
  //
  // sig must be exactly 64 bytes (IEEE P1363 raw r ‖ s, each 32
  // bytes). The leaf's PublicKey must be *ecdsa.PublicKey on a
  // P-256 curve. Any other shape fails with ReasonSignature.
  func verifySignature(leaf *x509.Certificate, headerB64, payloadB64 string, sig []byte) error {
      if len(sig) != 64 {
          return &VerificationError{
              Reason: ReasonSignature,
              Cause:  errors.New("signature: expected 64 bytes (ES256 IEEE P1363)"),
          }
      }
      pub, ok := leaf.PublicKey.(*ecdsa.PublicKey)
      if !ok {
          return &VerificationError{
              Reason: ReasonSignature,
              Cause:  errors.New("signature: leaf public key is not ECDSA"),
          }
      }
      hash := sha256.Sum256([]byte(headerB64 + "." + payloadB64))
      r := new(big.Int).SetBytes(sig[:32])
      s := new(big.Int).SetBytes(sig[32:])
      if !ecdsa.Verify(pub, hash[:], r, s) {
          return &VerificationError{
              Reason: ReasonSignature,
              Cause:  errors.New("signature: ECDSA verify failed"),
          }
      }
      return nil
  }
  ```

- [ ] **Step 4: Run, verify pass**

  ```bash
  go test -race ./jws/
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add jws/chain.go jws/chain_test.go
  git commit -m "feat(jws): ES256 signature verification (raw 64-byte only)"
  ```

---

## Task 9: `Verifier` struct + `NewVerifier` + Options

**Files:**
- Create: `jws/verifier.go`
- Create: `jws/verifier_test.go`

- [ ] **Step 1: Write failing tests**

  Write `jws/verifier_test.go`:
  ```go
  package jws

  import (
      "crypto/x509"
      "encoding/asn1"
      "testing"
      "time"
  )

  func TestNewVerifier_Defaults(t *testing.T) {
      v := NewVerifier()
      if v.roots == nil {
          t.Fatalf("expected non-nil default roots pool (will be set by DefaultVerifier)")
      }
      if len(v.requiredOIDs) == 0 {
          t.Fatalf("expected default required OIDs to be set")
      }
      if v.clock == nil {
          t.Fatalf("expected default clock function")
      }
  }

  func TestNewVerifier_WithRootCAs(t *testing.T) {
      pool := x509.NewCertPool()
      v := NewVerifier(WithRootCAs(pool))
      if v.roots != pool {
          t.Fatalf("WithRootCAs did not set the pool")
      }
  }

  func TestNewVerifier_WithRequiredOIDs(t *testing.T) {
      myOID := asn1.ObjectIdentifier{1, 2, 3}
      v := NewVerifier(WithRequiredOIDs(myOID))
      if len(v.requiredOIDs) != 1 || !v.requiredOIDs[0].Equal(myOID) {
          t.Fatalf("WithRequiredOIDs did not replace OID list")
      }
  }

  func TestNewVerifier_WithClock(t *testing.T) {
      fixed := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
      v := NewVerifier(WithClock(func() time.Time { return fixed }))
      if !v.clock().Equal(fixed) {
          t.Fatalf("WithClock did not install custom clock")
      }
  }
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: undefined references.

- [ ] **Step 3: Implement `verifier.go`**

  Write `jws/verifier.go`:
  ```go
  package jws

  import (
      "crypto/x509"
      "encoding/asn1"
      "time"
  )

  // Verifier holds the trust anchors, required OIDs, and clock used
  // by VerifyAndDecode. Construct with NewVerifier or grab the
  // process-wide DefaultVerifier.
  //
  // Once constructed, a Verifier is safe for concurrent use by
  // multiple goroutines. Its fields are not modified after
  // construction.
  type Verifier struct {
      roots        *x509.CertPool
      requiredOIDs []asn1.ObjectIdentifier
      clock        func() time.Time
  }

  // Option mutates a Verifier during NewVerifier.
  type Option func(*Verifier)

  // NewVerifier builds a Verifier. With no options it uses
  // package defaults: an empty roots pool (callers MUST supply roots
  // via WithRootCAs or use DefaultVerifier), DefaultRequiredOIDs as
  // the required OID list, and time.Now as the clock.
  //
  // For production code, prefer DefaultVerifier(), which embeds
  // Apple Root CA G3 automatically.
  func NewVerifier(opts ...Option) *Verifier {
      v := &Verifier{
          roots:        x509.NewCertPool(),
          requiredOIDs: append([]asn1.ObjectIdentifier(nil), DefaultRequiredOIDs...),
          clock:        time.Now,
      }
      for _, opt := range opts {
          opt(v)
      }
      return v
  }

  // WithRootCAs replaces the trust anchor pool. The Verifier takes
  // the pool as-is; do not mutate it after construction.
  func WithRootCAs(pool *x509.CertPool) Option {
      return func(v *Verifier) { v.roots = pool }
  }

  // WithRequiredOIDs replaces the required OID list. A leaf cert
  // passes if it carries ANY of the listed OIDs (logical OR).
  func WithRequiredOIDs(oids ...asn1.ObjectIdentifier) Option {
      return func(v *Verifier) {
          v.requiredOIDs = append([]asn1.ObjectIdentifier(nil), oids...)
      }
  }

  // WithClock replaces the verification clock. Tests use this to
  // exercise expired / not-yet-valid certificate paths
  // deterministically; production code should not call it.
  func WithClock(now func() time.Time) Option {
      return func(v *Verifier) { v.clock = now }
  }
  ```

- [ ] **Step 4: Run, verify pass**

  ```bash
  go test -race ./jws/
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add jws/verifier.go jws/verifier_test.go
  git commit -m "feat(jws): Verifier struct + NewVerifier + Option pattern"
  ```

---

## Task 10: `VerifyAndDecode[T]` integration (full test matrix)

This is the public entry point. It composes `X5c.Parse`, `verifyChain`, `matchOID`, `verifySignature`, plus header / payload JSON decoding. The test matrix here is the spec §8.3 table.

**Files:**
- Create: `jws/decode.go`
- Create: `jws/decode_test.go`

- [ ] **Step 1: Write failing tests covering the full matrix**

  Write `jws/decode_test.go`:
  ```go
  package jws

  import (
      "crypto/x509"
      "encoding/asn1"
      "encoding/base64"
      "encoding/json"
      "strings"
      "testing"
      "time"

      "github.com/godrealms/go-apple-sdk/jws/internal/testchain"
  )

  type tp struct {
      Foo string `json:"foo"`
      Bar int    `json:"bar"`
  }

  // newTestVerifier returns a Verifier configured to trust the chain's
  // root and require the chain's stamped OID.
  func newTestVerifier(tc *testchain.Chain) *Verifier {
      return NewVerifier(
          WithRootCAs(tc.RootPool),
          WithRequiredOIDs(OIDAppleReceiptSigning),
      )
  }

  // 1. Happy path
  func TestVerifyAndDecode_Success(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, tp{Foo: "hello", Bar: 42})
      out, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
      if err != nil {
          t.Fatalf("expected success, got %v", err)
      }
      if out.Foo != "hello" || out.Bar != 42 {
          t.Fatalf("decoded payload = %+v, want {hello 42}", out)
      }
  }

  // 2. Wrong root
  func TestVerifyAndDecode_UnknownRoot(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, tp{})
      v := NewVerifier(
          WithRootCAs(x509.NewCertPool()), // empty
          WithRequiredOIDs(OIDAppleReceiptSigning),
      )
      _, err := VerifyAndDecode[tp](v, raw)
      assertReason(t, err, ReasonChain)
  }

  // 3. Leaf expired (NotAfter in the past)
  func TestVerifyAndDecode_LeafExpired(t *testing.T) {
      tc := testchain.New(t, testchain.WithLeafNotAfter(time.Now().Add(-time.Hour)))
      raw := tc.SignJWS(t, tp{})
      _, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
      assertReason(t, err, ReasonExpired)
  }

  // 4. Leaf not yet valid (NotBefore in the future)
  func TestVerifyAndDecode_LeafNotYetValid(t *testing.T) {
      tc := testchain.New(t, testchain.WithLeafNotBefore(time.Now().Add(time.Hour)))
      raw := tc.SignJWS(t, tp{})
      _, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
      assertReason(t, err, ReasonExpired)
  }

  // 5. Leaf missing required OID
  func TestVerifyAndDecode_MissingOID(t *testing.T) {
      // Stamp leaf with an unrelated OID.
      tc := testchain.New(t, testchain.WithLeafOIDs(asn1.ObjectIdentifier{1, 2, 3, 4}))
      raw := tc.SignJWS(t, tp{})
      _, err := VerifyAndDecode[tp](newTestVerifier(tc), raw)
      assertReason(t, err, ReasonOID)
  }

  // 6. Tampered payload
  func TestVerifyAndDecode_TamperedPayload(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, tp{Foo: "x"})
      bits := strings.SplitN(raw, ".", 3)
      bits[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"foo":"y"}`))
      tampered := strings.Join(bits, ".")
      _, err := VerifyAndDecode[tp](newTestVerifier(tc), tampered)
      assertReason(t, err, ReasonSignature)
  }

  // 7. Truncated signature
  func TestVerifyAndDecode_TruncatedSignature(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, tp{})
      bits := strings.SplitN(raw, ".", 3)
      sig, _ := base64.RawURLEncoding.DecodeString(bits[2])
      bits[2] = base64.RawURLEncoding.EncodeToString(sig[:60])
      mangled := strings.Join(bits, ".")
      _, err := VerifyAndDecode[tp](newTestVerifier(tc), mangled)
      assertReason(t, err, ReasonSignature)
  }

  // 8. alg != ES256
  func TestVerifyAndDecode_WrongAlg(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, tp{})
      bits := strings.SplitN(raw, ".", 3)
      hdr, _ := base64.RawURLEncoding.DecodeString(bits[0])
      var hh map[string]any
      _ = json.Unmarshal(hdr, &hh)
      hh["alg"] = "RS256"
      newHdr, _ := json.Marshal(hh)
      bits[0] = base64.RawURLEncoding.EncodeToString(newHdr)
      mangled := strings.Join(bits, ".")
      _, err := VerifyAndDecode[tp](newTestVerifier(tc), mangled)
      assertReason(t, err, ReasonStructure)
  }

  // 9. Wrong segment count
  func TestVerifyAndDecode_TwoSegments(t *testing.T) {
      _, err := VerifyAndDecode[tp](NewVerifier(), "abc.def")
      assertReason(t, err, ReasonStructure)
  }

  // 10. Empty x5c
  func TestVerifyAndDecode_EmptyX5c(t *testing.T) {
      hdr, _ := json.Marshal(Header{Alg: "ES256", X5c: X5c{}})
      payload, _ := json.Marshal(tp{})
      raw := base64.RawURLEncoding.EncodeToString(hdr) + "." +
          base64.RawURLEncoding.EncodeToString(payload) + "." +
          base64.RawURLEncoding.EncodeToString(make([]byte, 64))
      _, err := VerifyAndDecode[tp](NewVerifier(), raw)
      assertReason(t, err, ReasonStructure)
  }

  // 11. Bad base64 inside x5c
  func TestVerifyAndDecode_BadX5cBase64(t *testing.T) {
      hdr, _ := json.Marshal(Header{Alg: "ES256", X5c: X5c{"!!!not-b64!!!"}})
      payload, _ := json.Marshal(tp{})
      raw := base64.RawURLEncoding.EncodeToString(hdr) + "." +
          base64.RawURLEncoding.EncodeToString(payload) + "." +
          base64.RawURLEncoding.EncodeToString(make([]byte, 64))
      _, err := VerifyAndDecode[tp](NewVerifier(), raw)
      assertReason(t, err, ReasonStructure)
  }

  // 12. RSA leaf — covered by TestVerifySignature_NonECDSAKey in
  // chain_test.go. No duplication here.

  // 13. Real Apple notification regression — added in Task 18.
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: `undefined: VerifyAndDecode`.

- [ ] **Step 3: Implement `decode.go`**

  Write `jws/decode.go`:
  ```go
  package jws

  import (
      "encoding/base64"
      "encoding/json"
      "errors"
      "fmt"
      "strings"
  )

  // VerifyAndDecode is the single verification entry point. It
  // splits the JWS, validates the certificate chain, checks the leaf
  // OID, verifies the ES256 signature, and only then decodes the
  // payload as *T. Any failure returns *VerificationError; see
  // ReasonCode for the categories.
  //
  // The signature is verified BEFORE the payload is JSON-decoded so
  // we never run a JSON parser on untrusted bytes.
  func VerifyAndDecode[T any](v *Verifier, raw string) (*T, error) {
      // 1. Split into 3 segments.
      segments := strings.SplitN(raw, ".", 3)
      if len(segments) != 3 {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("expected 3 JWS segments, got %d", len(segments)),
          }
      }
      headerB64, payloadB64, sigB64 := segments[0], segments[1], segments[2]

      // 2. Decode header; enforce alg=ES256.
      headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
      if err != nil {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("header base64: %w", err),
          }
      }
      var header Header
      if err := json.Unmarshal(headerBytes, &header); err != nil {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("header json: %w", err),
          }
      }
      if header.Alg != "ES256" {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("unsupported alg %q (only ES256)", header.Alg),
          }
      }

      // 3. Parse x5c chain.
      chain, err := header.X5c.Parse() // already returns *VerificationError
      if err != nil {
          return nil, err
      }

      // 4. Chain validation.
      if err := verifyChain(chain, v.roots, v.clock()); err != nil {
          return nil, err
      }

      // 5. OID check.
      if !matchOID(chain[0].Extensions, v.requiredOIDs) {
          return nil, &VerificationError{
              Reason: ReasonOID,
              Cause:  errors.New("leaf cert carries none of the required OIDs"),
          }
      }

      // 6. Signature verification — BEFORE payload JSON decode.
      sig, err := base64.RawURLEncoding.DecodeString(sigB64)
      if err != nil {
          return nil, &VerificationError{
              Reason: ReasonSignature,
              Cause:  fmt.Errorf("signature base64: %w", err),
          }
      }
      if err := verifySignature(chain[0], headerB64, payloadB64, sig); err != nil {
          return nil, err
      }

      // 7. Decode payload (only after signature passes).
      payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
      if err != nil {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("payload base64: %w", err),
          }
      }
      out := new(T)
      if err := json.Unmarshal(payloadBytes, out); err != nil {
          return nil, &VerificationError{
              Reason: ReasonStructure,
              Cause:  fmt.Errorf("payload json: %w", err),
          }
      }
      return out, nil
  }
  ```

- [ ] **Step 4: Run all `jws/` tests, verify pass**

  ```bash
  go test -race -cover ./jws/
  ```

  Expected: PASS, coverage ≥ 90%.

- [ ] **Step 5: Commit**

  ```bash
  git add jws/decode.go jws/decode_test.go
  git commit -m "feat(jws): VerifyAndDecode[T] integration + full failure matrix"
  ```

---

## Task 11: Embed Apple Root CA G3 + `DefaultVerifier`

**Files:**
- Create: `jws/apple_root_ca_g3.pem`
- Create: `jws/default.go`
- Create: `jws/default_test.go`

- [ ] **Step 1: Download Apple Root CA G3 and verify integrity**

  ```bash
  mkdir -p /tmp/apple-ca
  curl -fsSL https://www.apple.com/certificateauthority/AppleRootCA-G3.cer \
      -o /tmp/apple-ca/AppleRootCA-G3.cer

  # Convert DER → PEM.
  openssl x509 -inform DER -in /tmp/apple-ca/AppleRootCA-G3.cer \
      -out /tmp/apple-ca/AppleRootCA-G3.pem

  # Print SHA-256 of the DER bytes (this is the well-known
  # fingerprint to record).
  openssl dgst -sha256 /tmp/apple-ca/AppleRootCA-G3.cer
  ```

  Expected fingerprint (publicly published by Apple — Subject Key Identifier `bb b0 dd 7e 76 56 6c c9 e6 4c e7 9c b1 ad 4d 1d 88 96 5d a4`; SHA-256 of DER `63 34 3a bf b8 9a 6a 03 eb b5 7e 9b 3f 5f a7 be 7c 4f 5c 75 6f 30 17 b3 a8 c4 88 c3 65 3e 91 79`). If your downloaded file's SHA-256 disagrees, **stop** — do not embed an unknown cert.

  Move the PEM into the package:
  ```bash
  cp /tmp/apple-ca/AppleRootCA-G3.pem \
      /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk/jws/apple_root_ca_g3.pem
  ```

- [ ] **Step 2: Write failing test for `DefaultVerifier`**

  Write `jws/default_test.go`:
  ```go
  package jws

  import (
      "testing"
  )

  func TestDefaultVerifier_LoadsEmbeddedRoot(t *testing.T) {
      v := DefaultVerifier()
      if v == nil {
          t.Fatalf("DefaultVerifier returned nil")
      }
      if v.roots == nil {
          t.Fatalf("DefaultVerifier root pool is nil")
      }
      // CertPool exposes Subjects() which returns DER-encoded
      // RDNSequences. We just need at least one — embedded root.
      if len(v.roots.Subjects()) == 0 {
          t.Fatalf("DefaultVerifier root pool is empty")
      }
  }

  func TestDefaultVerifier_Singleton(t *testing.T) {
      a := DefaultVerifier()
      b := DefaultVerifier()
      if a != b {
          t.Fatalf("DefaultVerifier should return the same instance across calls")
      }
  }
  ```

- [ ] **Step 3: Run, verify failure**

  ```bash
  go test ./jws/
  ```

  Expected: `undefined: DefaultVerifier`.

- [ ] **Step 4: Implement `default.go`**

  Write `jws/default.go`:
  ```go
  package jws

  import (
      _ "embed"
      "crypto/x509"
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
  ```

- [ ] **Step 5: Run, verify pass**

  ```bash
  go test -race ./jws/
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add jws/apple_root_ca_g3.pem jws/default.go jws/default_test.go
  git commit -m "feat(jws): embed Apple Root CA G3 + DefaultVerifier (sync.Once)"
  ```

---

## Task 12: Concurrent stress test for `DefaultVerifier`

**Files:**
- Modify: `jws/default_test.go`

- [ ] **Step 1: Append concurrent test**

  Append to `jws/default_test.go`:
  ```go
  import "sync"

  func TestDefaultVerifier_ConcurrentAccess(t *testing.T) {
      const n = 100
      var wg sync.WaitGroup
      verifiers := make([]*Verifier, n)
      wg.Add(n)
      for i := 0; i < n; i++ {
          go func(i int) {
              defer wg.Done()
              verifiers[i] = DefaultVerifier()
          }(i)
      }
      wg.Wait()
      for i := 1; i < n; i++ {
          if verifiers[i] != verifiers[0] {
              t.Fatalf("verifier[%d] differs from verifier[0]", i)
          }
      }
  }
  ```

- [ ] **Step 2: Run with race detector**

  ```bash
  go test -race ./jws/
  ```

  Expected: PASS (no data race report).

- [ ] **Step 3: Commit**

  ```bash
  git add jws/default_test.go
  git commit -m "test(jws): concurrent DefaultVerifier access stress test"
  ```

---

## Task 13: `scripts/update-root-ca.sh`

**Files:**
- Create: `scripts/update-root-ca.sh`

- [ ] **Step 1: Write the script**

  ```bash
  mkdir -p /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk/scripts
  ```

  Write `scripts/update-root-ca.sh`:
  ```bash
  #!/usr/bin/env bash
  # Refresh jws/apple_root_ca_g3.pem from Apple's published source.
  # Verifies the SHA-256 of the downloaded DER bytes before writing.
  #
  # Run from repo root: ./scripts/update-root-ca.sh
  set -euo pipefail

  URL="https://www.apple.com/certificateauthority/AppleRootCA-G3.cer"
  EXPECTED_SHA256="63343abfb89a6a03ebb57e9b3f5fa7be7c4f5c756f3017b3a8c488c3653e9179"
  TARGET="jws/apple_root_ca_g3.pem"

  if [[ ! -d jws ]]; then
      echo "Run from repo root (jws/ directory not found)." >&2
      exit 1
  fi

  TMP=$(mktemp -d)
  trap 'rm -rf "$TMP"' EXIT

  echo "Fetching $URL ..."
  curl -fsSL "$URL" -o "$TMP/cert.cer"

  ACTUAL=$(shasum -a 256 "$TMP/cert.cer" | awk '{print $1}')
  if [[ "$ACTUAL" != "$EXPECTED_SHA256" ]]; then
      echo "SHA-256 mismatch."
      echo "  expected: $EXPECTED_SHA256"
      echo "  actual:   $ACTUAL"
      echo "REFUSING to overwrite $TARGET. If Apple legitimately"
      echo "rotated the root cert, update EXPECTED_SHA256 in this"
      echo "script after independent verification."
      exit 1
  fi

  openssl x509 -inform DER -in "$TMP/cert.cer" -out "$TMP/cert.pem"
  mv "$TMP/cert.pem" "$TARGET"
  echo "Updated $TARGET (sha256=$ACTUAL)."
  ```

- [ ] **Step 2: Make executable**

  ```bash
  chmod +x /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk/scripts/update-root-ca.sh
  ```

- [ ] **Step 3: Run it (idempotent — should produce no diff)**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk
  ./scripts/update-root-ca.sh
  git diff jws/apple_root_ca_g3.pem
  ```

  Expected: script prints "Updated …", `git diff` is empty (the embedded PEM matches what the script just downloaded).

- [ ] **Step 4: Commit**

  ```bash
  git add scripts/update-root-ca.sh
  git commit -m "chore(scripts): update-root-ca.sh refreshes embedded G3 PEM"
  ```

---

## Task 14: Migrate `types/JWSTransaction.Decrypt`

**Files:**
- Modify: `types/JWSTransaction.go`

- [ ] **Step 1: Write the new test**

  Create `types/JWSTransaction_test.go` (file does not exist yet):
  ```go
  package types

  import (
      "testing"

      "github.com/godrealms/go-apple-sdk/jws"
      "github.com/godrealms/go-apple-sdk/jws/internal/testchain"
  )

  func TestJWSTransaction_Decrypt_UsesDefaultVerifier(t *testing.T) {
      // The default Verifier requires the real Apple Root CA, so a
      // testchain payload will fail with ReasonChain. That's the
      // signal we want: it proves Decrypt is now hitting the
      // verifier path (rather than blindly accepting any cert as
      // before).
      tc := testchain.New(t)
      raw := tc.SignJWS(t, JWSTransactionDecodedPayload{TransactionId: "abc"})
      _, err := JWSTransaction(raw).Decrypt()
      ve, ok := err.(*jws.VerificationError)
      if !ok {
          t.Fatalf("expected *jws.VerificationError, got %T: %v", err, err)
      }
      if ve.Reason != jws.ReasonChain {
          t.Fatalf("expected ReasonChain, got %s", ve.Reason)
      }
  }

  func TestJWSTransaction_DecryptWith_AcceptsTestVerifier(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, JWSTransactionDecodedPayload{TransactionId: "tx-123"})
      v := jws.NewVerifier(
          jws.WithRootCAs(tc.RootPool),
          jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning),
      )
      out, err := JWSTransaction(raw).DecryptWith(v)
      if err != nil {
          t.Fatalf("expected success, got %v", err)
      }
      if string(out.TransactionId) != "tx-123" {
          t.Fatalf("transaction id = %q, want tx-123", out.TransactionId)
      }
  }
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./types/
  ```

  Expected: `Decrypt undefined` errors disappear (Decrypt exists), but `DecryptWith` is undefined.

- [ ] **Step 3: Rewrite `Decrypt`, add `DecryptWith`, delete old code**

  Open `types/JWSTransaction.go` and **replace lines 1–192** (the entire file except the `JWSTransaction` and `JWSTransactionDecodedPayload` type definitions) with:

  Final shape of the file:
  ```go
  package types

  import "github.com/godrealms/go-apple-sdk/jws"

  // (keep all the existing private typedefs: currency, isUpgraded,
  //  offerDiscountType, offerIdentifier, offerType, price, quantity,
  //  revocationReason, storefront, storefrontId, transactionReason)
  //
  // (keep JWSTransactionDecodedPayload struct unchanged)

  // JWSTransaction is the JWS-Compact-Serialised transaction Apple
  // returns from App Store Server API endpoints. Decrypt verifies
  // and decodes it using the package-default Verifier. To override
  // the trust anchors (e.g. for tests), use DecryptWith.
  type JWSTransaction string

  // Decrypt verifies the JWS chain + signature and returns the
  // decoded payload. Returns *jws.VerificationError on failure.
  func (j JWSTransaction) Decrypt() (*JWSTransactionDecodedPayload, error) {
      return jws.VerifyAndDecode[JWSTransactionDecodedPayload](jws.DefaultVerifier(), string(j))
  }

  // DecryptWith verifies using the supplied Verifier instead of the
  // package default.
  func (j JWSTransaction) DecryptWith(v *jws.Verifier) (*JWSTransactionDecodedPayload, error) {
      return jws.VerifyAndDecode[JWSTransactionDecodedPayload](v, string(j))
  }
  ```

  Remove all imports that are no longer used: `crypto`, `crypto/ecdsa`, `crypto/rsa`, `crypto/sha256`, `encoding/base64`, `fmt`, `math/big`, `strings`, `encoding/json`. Keep only `github.com/godrealms/go-apple-sdk/jws`.

  Delete the old `parseSignedPayload` method entirely.

- [ ] **Step 4: Run tests, verify pass + no compile breaks elsewhere**

  ```bash
  go vet ./...
  go test -race ./types/ ./app-store-server/...
  ```

  Expected: PASS.

- [ ] **Step 5: Commit**

  ```bash
  git add types/JWSTransaction.go types/JWSTransaction_test.go
  git commit -m "refactor(types): JWSTransaction.Decrypt uses jws.VerifyAndDecode

  Adds DecryptWith(*jws.Verifier) overload for callers that want
  custom trust anchors. Deletes the inline (and broken)
  parseSignedPayload + ecdsa/rsa fallback path — the new path
  goes through jws.DefaultVerifier which validates the cert chain."
  ```

---

## Task 15: Migrate `types/JWSRenewalInfo.Decrypt`

Same pattern as Task 14.

**Files:**
- Modify: `types/JWSRenewalInfo.go`
- Create: `types/JWSRenewalInfo_test.go`

- [ ] **Step 1: Write test**

  Write `types/JWSRenewalInfo_test.go`:
  ```go
  package types

  import (
      "testing"

      "github.com/godrealms/go-apple-sdk/jws"
      "github.com/godrealms/go-apple-sdk/jws/internal/testchain"
  )

  func TestJWSRenewalInfo_Decrypt_UsesDefaultVerifier(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, JWSRenewalInfoDecodedPayload{ProductId: "p"})
      _, err := JWSRenewalInfo(raw).Decrypt()
      ve, ok := err.(*jws.VerificationError)
      if !ok || ve.Reason != jws.ReasonChain {
          t.Fatalf("expected ReasonChain, got %v", err)
      }
  }

  func TestJWSRenewalInfo_DecryptWith_AcceptsTestVerifier(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, JWSRenewalInfoDecodedPayload{ProductId: "p-9"})
      v := jws.NewVerifier(jws.WithRootCAs(tc.RootPool),
          jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning))
      out, err := JWSRenewalInfo(raw).DecryptWith(v)
      if err != nil {
          t.Fatalf("expected success, got %v", err)
      }
      if string(out.ProductId) != "p-9" {
          t.Fatalf("product id = %q", out.ProductId)
      }
  }
  ```

- [ ] **Step 2: Verify failure (DecryptWith undefined)**

  ```bash
  go test ./types/
  ```

- [ ] **Step 3: Replace `JWSRenewalInfo.go` body**

  Same shape as Task 14: keep type definitions; delete `Decrypt` body + `parseSignedPayload`; rewrite as 2 thin methods that delegate to `jws.VerifyAndDecode`.

  ```go
  package types

  import "github.com/godrealms/go-apple-sdk/jws"

  // (keep private typedefs: autoRenewProductId, autoRenewStatus,
  //  eligibleWinBackOfferIds, expirationIntent, isInBillingRetryPeriod,
  //  priceIncreaseStatus, renewalPrice)

  // (keep JWSRenewalInfoDecodedPayload struct unchanged)

  // JWSRenewalInfo is JWS-encoded subscription renewal info from
  // Apple. Decrypt verifies + decodes using DefaultVerifier; use
  // DecryptWith for custom trust anchors.
  type JWSRenewalInfo string

  func (j JWSRenewalInfo) Decrypt() (*JWSRenewalInfoDecodedPayload, error) {
      return jws.VerifyAndDecode[JWSRenewalInfoDecodedPayload](jws.DefaultVerifier(), string(j))
  }

  func (j JWSRenewalInfo) DecryptWith(v *jws.Verifier) (*JWSRenewalInfoDecodedPayload, error) {
      return jws.VerifyAndDecode[JWSRenewalInfoDecodedPayload](v, string(j))
  }
  ```

  Trim imports the same way.

- [ ] **Step 4: Verify pass**

  ```bash
  go vet ./...
  go test -race ./types/ ./app-store-server/...
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add types/JWSRenewalInfo.go types/JWSRenewalInfo_test.go
  git commit -m "refactor(types): JWSRenewalInfo.Decrypt uses jws.VerifyAndDecode"
  ```

---

## Task 16: Migrate `app-store-server-notifications/V2.DecodedPayload`

Same pattern as Tasks 14 and 15.

**Files:**
- Modify: `app-store-server-notifications/App.Store.Server.Notifications.V2.go`
- Create: `app-store-server-notifications/v2_test.go`

- [ ] **Step 1: Write test**

  Write `app-store-server-notifications/v2_test.go`:
  ```go
  package AppStoreNotifications

  import (
      "testing"

      "github.com/godrealms/go-apple-sdk/jws"
      "github.com/godrealms/go-apple-sdk/jws/internal/testchain"
  )

  func TestSignedPayload_DecodedPayload_UsesDefaultVerifier(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, ResponseBodyV2DecodedPayload{NotificationUUID: "u-1"})
      _, err := SignedPayload(raw).DecodedPayload()
      ve, ok := err.(*jws.VerificationError)
      if !ok || ve.Reason != jws.ReasonChain {
          t.Fatalf("expected ReasonChain, got %v", err)
      }
  }

  func TestSignedPayload_DecodedPayloadWith_AcceptsTestVerifier(t *testing.T) {
      tc := testchain.New(t)
      raw := tc.SignJWS(t, ResponseBodyV2DecodedPayload{NotificationUUID: "u-42"})
      v := jws.NewVerifier(jws.WithRootCAs(tc.RootPool),
          jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning))
      out, err := SignedPayload(raw).DecodedPayloadWith(v)
      if err != nil {
          t.Fatalf("expected success, got %v", err)
      }
      if string(out.NotificationUUID) != "u-42" {
          t.Fatalf("notification uuid = %q", out.NotificationUUID)
      }
  }
  ```

- [ ] **Step 2: Run, verify failure**

  ```bash
  go test ./app-store-server-notifications/...
  ```

- [ ] **Step 3: Replace `DecodedPayload` body + delete `parseSignedPayload`**

  Open `app-store-server-notifications/App.Store.Server.Notifications.V2.go`. **Replace lines 96–178** with:

  ```go
  type SignedPayload string

  // DecodedPayload verifies the JWS chain + signature and returns
  // the decoded notification payload. Returns *jws.VerificationError
  // on failure.
  func (sp SignedPayload) DecodedPayload() (*ResponseBodyV2DecodedPayload, error) {
      return jws.VerifyAndDecode[ResponseBodyV2DecodedPayload](
          jws.DefaultVerifier(), string(sp))
  }

  // DecodedPayloadWith verifies using the supplied Verifier.
  func (sp SignedPayload) DecodedPayloadWith(v *jws.Verifier) (*ResponseBodyV2DecodedPayload, error) {
      return jws.VerifyAndDecode[ResponseBodyV2DecodedPayload](v, string(sp))
  }
  ```

  Add the import:
  ```go
  import "github.com/godrealms/go-apple-sdk/jws"
  ```

  Remove now-unused imports: `crypto`, `crypto/ecdsa`, `crypto/rsa`, `crypto/sha256`, `encoding/base64`, `math/big`, `strings`. Keep `encoding/json`, `fmt` if still needed by other code in the file (they're not, after this change — drop them too).

- [ ] **Step 4: Verify pass**

  ```bash
  go vet ./...
  go test -race ./app-store-server-notifications/...
  ```

- [ ] **Step 5: Commit**

  ```bash
  git add app-store-server-notifications/
  git commit -m "refactor(notifications): V2 DecodedPayload uses jws.VerifyAndDecode"
  ```

---

## Task 17: Collapse `types/JWSDecodedHeader.go` to aliases + delete `types/x5c.go`

**Files:**
- Modify: `types/JWSDecodedHeader.go`
- Delete: `types/x5c.go`

- [ ] **Step 1: Confirm no internal references**

  ```bash
  cd /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk
  grep -rn "types\.X5c\|types\.JWSDecodedHeader\|header\.X5c\|header\.Alg" \
      --include="*.go" .
  ```

  Expected: only references inside `types/JWSDecodedHeader.go` itself and possibly a few aliases — no consumers outside the file. If any consumer exists, it must already be dead after Tasks 14–16 (parseSignedPayload was the only consumer).

- [ ] **Step 2: Replace `types/JWSDecodedHeader.go`**

  Overwrite the file with:
  ```go
  package types

  import "github.com/godrealms/go-apple-sdk/jws"

  // Alg is a JWS "alg" header value, aliased from the jws package.
  // Kept here so existing imports of types.Alg continue to compile.
  type Alg = jws.Alg

  // X5c is the JWS x5c header field, aliased from the jws package.
  //
  // The previous types.X5c.GetPublicKey() method has been removed.
  // It returned only the leaf cert without validating the chain,
  // which made every consumer trivially impersonable. Migrate to
  // jws.Verifier (or the new Decrypt / DecodedPayload methods on
  // JWSTransaction / JWSRenewalInfo / SignedPayload, which all use
  // jws.DefaultVerifier internally).
  type X5c = jws.X5c

  // JWSDecodedHeader is the JWS protected header, aliased from the
  // jws package.
  type JWSDecodedHeader = jws.Header
  ```

- [ ] **Step 3: Delete `types/x5c.go`**

  ```bash
  git rm types/x5c.go
  ```

- [ ] **Step 4: Verify build + tests still green**

  ```bash
  go vet ./...
  go test -race ./...
  ```

  Expected: no compile errors, all tests pass.

- [ ] **Step 5: Commit**

  ```bash
  git add types/JWSDecodedHeader.go
  git commit -m "refactor(types): JWSDecodedHeader/X5c/Alg become jws aliases

  Deletes types/x5c.go (orphan duplicate, never used internally).
  X5c.GetPublicKey() is intentionally removed — it skipped chain
  validation, which is the whole bug we're fixing. Callers that
  used it must migrate to jws.Verifier."
  ```

---

## Task 18: Real Apple sandbox notification regression fixture

**Files:**
- Create: `jws/testdata/real_apple_notification.txt`
- Modify: `jws/decode_test.go`

- [ ] **Step 1: Capture a notification + redact PII**

  Get a real V2 notification's `signedPayload` from a sandbox run (same as Task 0 Step 1). Redact any PII in the payload — but the JWS string is not editable without breaking the signature, so the better approach is:

  - Use a notification that does NOT contain user-identifiable data (a `TEST` notification type emits a minimal payload).
  - Save the entire JWS string (single line, no trailing newline) to `jws/testdata/real_apple_notification.txt`.

- [ ] **Step 2: Append regression test to `jws/decode_test.go`**

  ```go
  import (
      _ "embed"
      // ...existing imports...
  )

  //go:embed testdata/real_apple_notification.txt
  var realAppleNotification string

  // type RealNotificationPayload mirrors the bits we care about for
  // the regression test. Apple sends many more fields; encoding/json
  // ignores unknowns.
  type realNotificationPayload struct {
      NotificationUUID string `json:"notificationUUID"`
      NotificationType string `json:"notificationType"`
  }

  func TestVerifyAndDecode_RealAppleNotification(t *testing.T) {
      if realAppleNotification == "" {
          t.Skip("real_apple_notification.txt not populated yet")
      }
      out, err := VerifyAndDecode[realNotificationPayload](
          DefaultVerifier(),
          strings.TrimSpace(realAppleNotification),
      )
      if err != nil {
          t.Fatalf("real notification failed: %v", err)
      }
      if out.NotificationUUID == "" {
          t.Fatalf("notification UUID empty after decode")
      }
  }
  ```

- [ ] **Step 3: Run the test**

  ```bash
  go test -race -run TestVerifyAndDecode_RealAppleNotification ./jws/
  ```

  Expected: PASS — proves DefaultVerifier accepts a real Apple-signed payload end-to-end. (Skips if the fixture file is empty.)

- [ ] **Step 4: Commit**

  ```bash
  git add jws/testdata/real_apple_notification.txt jws/decode_test.go
  git commit -m "test(jws): real Apple sandbox notification regression lock"
  ```

---

## Task 19: README + CHANGELOG release notes

**Files:**
- Modify: `README.md` (Section: 错误处理 / Examples)
- Create: `CHANGELOG.md` (if not exists)

- [ ] **Step 1: Write CHANGELOG entry**

  Create `/Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk/CHANGELOG.md` (or prepend if it exists):
  ```markdown
  # Changelog

  ## Unreleased

  ### Security

  - **JWS chain validation**: All JWS verification paths now perform
    full RFC 5280 certificate chain validation against Apple Root CA
    G3 (embedded in the SDK) plus a mandatory leaf-cert OID check
    (Apple receipt-signing OID). Previously the SDK only verified
    the JWS signature itself — it accepted any leaf cert from any
    CA, which left every caller of `SignedPayload.DecodedPayload`,
    `JWSTransaction.Decrypt`, and `JWSRenewalInfo.Decrypt`
    impersonable. Affected: anyone consuming App Store Server
    Notifications V2 or signed transaction / renewal info from App
    Store Server API.

    **Action required:**

    - If you were relying on `Decrypt()` / `DecodedPayload()` to
      succeed on payloads NOT signed by Apple's real chain (e.g.
      self-signed test fixtures), pass a custom `*jws.Verifier`
      via the new `DecryptWith(v)` / `DecodedPayloadWith(v)`
      methods.

    - The deprecated `types.X5c.GetPublicKey()` method has been
      **removed**. It returned only the leaf cert without
      validating the chain. Code that called it will no longer
      compile. Migrate to `jws.Verifier`.

  ### Added

  - New top-level package `github.com/godrealms/go-apple-sdk/jws`
    with `*Verifier`, `VerifyAndDecode[T]`, and
    `*VerificationError` (Reason enum: structure / chain / oid /
    expired / signature).
  - `scripts/update-root-ca.sh` to refresh the embedded Apple Root
    CA G3 with SHA-256 verification.

  ### Changed

  - `JWSTransaction.Decrypt`, `JWSRenewalInfo.Decrypt`, and
    `SignedPayload.DecodedPayload` now return `*jws.VerificationError`
    on failure (still satisfies `error`; use `errors.As` to inspect
    the `Reason`). Existing code that only checked `err != nil`
    continues to work.

  ### Removed

  - `types/x5c.go` (orphaned duplicate of the X5c type — never
    referenced internally; consolidated into `jws.X5c`).
  - The RSA and ASN.1 signature-verification fallback paths in
    legacy `Decrypt` / `DecodedPayload` implementations. Apple's
    documented JWS profile is ES256-only.
  ```

- [ ] **Step 2: Add a "JWS 验签错误处理" subsection to README.md's 错误处理 section**

  Open `README.md` and insert after the existing `### 错误处理` section:
  ```markdown
  ### JWS 验签错误处理

  从 SDK v\<next\> 起，所有 JWS payload 解密路径（`SignedPayload.DecodedPayload`、`JWSTransaction.Decrypt`、`JWSRenewalInfo.Decrypt`）都会校验证书链回到 Apple Root CA G3。失败时返回 `*jws.VerificationError`：

  ```go
  import (
      "errors"

      "github.com/godrealms/go-apple-sdk/jws"
      AppStoreNotifications "github.com/godrealms/go-apple-sdk/app-store-server-notifications"
  )

  payload, err := AppStoreNotifications.SignedPayload(raw).DecodedPayload()
  if err != nil {
      var ve *jws.VerificationError
      if errors.As(err, &ve) {
          switch ve.Reason {
          case jws.ReasonExpired:
              // 证书过期 —— 通常是 SDK 该升级了
          case jws.ReasonChain:
              // 链验证失败 —— 可能是攻击或配置错
          case jws.ReasonStructure, jws.ReasonOID, jws.ReasonSignature:
              // 上游格式错或签名错
          }
      }
      return err
  }
  ```

  如果你在测试中需要使用自签证书（比如 mock Apple 通知），用 `DecodedPayloadWith` / `DecryptWith` 配合自定义 `*jws.Verifier`：

  ```go
  v := jws.NewVerifier(
      jws.WithRootCAs(myTestPool),
      jws.WithRequiredOIDs(jws.OIDAppleReceiptSigning),
  )
  payload, err := AppStoreNotifications.SignedPayload(raw).DecodedPayloadWith(v)
  ```
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add README.md CHANGELOG.md
  git commit -m "docs: README + CHANGELOG release notes for jws/ chain validation"
  ```

---

## Task 20: Update one example to demonstrate `*VerificationError` handling

**Files:**
- Modify: `examples/app-store-server-notifications/v2/main.go`

- [ ] **Step 1: Open the example and locate the existing decode**

  ```bash
  cat /Volumes/Fanxiang-S790-1TB-Media/Personal/sdk/go-apple-sdk/examples/app-store-server-notifications/v2/main.go
  ```

  Find the line that calls `.DecodedPayload()` or similar.

- [ ] **Step 2: Add `errors.As` branching**

  Replace whatever single-line error check exists with:
  ```go
  payload, err := signedPayload.DecodedPayload()
  if err != nil {
      var ve *jws.VerificationError
      if errors.As(err, &ve) {
          log.Printf("JWS verification failed: reason=%s, cause=%v", ve.Reason, ve.Cause)
          switch ve.Reason {
          case jws.ReasonChain, jws.ReasonExpired:
              log.Printf("  → cert chain problem; check whether SDK is up to date")
          case jws.ReasonOID:
              log.Printf("  → leaf cert missing required Apple OID")
          case jws.ReasonSignature:
              log.Printf("  → signature mismatch (tampered payload?)")
          case jws.ReasonStructure:
              log.Printf("  → upstream payload malformed")
          }
      }
      log.Fatal(err)
  }
  _ = payload
  ```

  Add the imports `errors` and `github.com/godrealms/go-apple-sdk/jws` if missing.

- [ ] **Step 3: Verify the example builds**

  ```bash
  go build ./examples/app-store-server-notifications/v2/...
  ```

  Expected: no errors.

- [ ] **Step 4: Final full-repo test sweep**

  ```bash
  go vet ./...
  go test -race -cover ./...
  ```

  Expected: all packages pass; `jws/` reports ≥ 90% coverage; no compile errors anywhere.

- [ ] **Step 5: Commit**

  ```bash
  git add examples/app-store-server-notifications/v2/main.go
  git commit -m "docs(examples): demo *jws.VerificationError handling in v2 notifications"
  ```

---

## Plan-level acceptance criteria

Before declaring this plan complete, all of the following must hold:

- [ ] Task 0 produced a recorded decision in `docs/superpowers/notes/2026-05-05-oid-confirmation.md`. If `6.29` was found absent, the spec and `jws/oid.go` `DefaultRequiredOIDs` were updated accordingly.
- [ ] `go vet ./...` clean.
- [ ] `go test -race ./...` all packages pass.
- [ ] `go test -cover ./jws/` reports ≥ 90% coverage.
- [ ] `git grep "parseSignedPayload"` returns nothing (all 3 copies deleted).
- [ ] `git grep "X5c.GetPublicKey"` returns nothing (method removed; no internal callers either).
- [ ] `git ls-files types/x5c.go` returns nothing (file deleted).
- [ ] `jws/apple_root_ca_g3.pem` exists and matches the SHA-256 in `scripts/update-root-ca.sh`.
- [ ] CHANGELOG entry written; README has the JWS 验签错误处理 subsection.

---

## Notes for the executing engineer

- **Run `Task 0` first.** Do not skip it. The whole plan's correctness pivots on the OID list. If you cannot get a sandbox notification, ask the project owner — do not guess.
- The `testchain` helper (Task 2) is the foundation for almost every test in this plan. Get it green before moving on; if `testchain_test.go` is flaky, every subsequent task will be flaky too.
- Keep tasks atomic — one commit per task. If you find yourself wanting to fix something unrelated, stop and add it to a TODO list for later (or use `mcp__ccd_session__spawn_task`).
- Per CONTRIBUTING / README: `golangci-lint run` is optional but recommended. The new code is intended to pass `revive`'s `exported` rule (every exported symbol has a comment).

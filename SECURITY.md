# Security Policy

`go-apple-sdk` performs security-sensitive work — JWS signature verification,
x509 certificate-chain validation, and App Store JWT signing — so we take
vulnerability reports seriously.

## Supported Versions

| Version | Supported                          |
| ------- | ---------------------------------- |
| latest  | ✅ Actively supported              |
| older   | ⚠️ Critical fixes on a best-effort basis |

## Reporting a Vulnerability

**Please do not open a public issue for security vulnerabilities.**

Use GitHub's [Private Vulnerability Reporting](https://github.com/godrealms/go-apple-sdk/security/advisories/new)
to report privately. We aim to acknowledge reports within 72 hours and will
coordinate a disclosure timeline with you before any fix is published.

When reporting, please include:

- The affected version(s) and a minimal reproduction if possible.
- The impact you believe the issue has (e.g. signature bypass, credential
  exposure, denial of service).

## Handling Credentials

This SDK signs requests with your App Store Connect private key (`.p8`).
Operational guidance:

- Never commit `.p8` / `.pem` / private-key material to source control. Inject
  it from an environment variable or a secrets manager at runtime.
- Do not pass externally-sourced, unvalidated identifiers directly into SDK
  calls; validate them first.
- The library redacts `Authorization` headers from its diagnostic error logs,
  but you should still avoid logging request/response bodies that may contain
  tokens in your own code.

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

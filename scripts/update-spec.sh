#!/usr/bin/env bash
# scripts/update-spec.sh
#
# Download Apple's App Store Connect OpenAPI spec, verify the SHA256,
# and copy the JSON payload into internal/cmd/gen-asc/spec/.
#
# Usage:
#   scripts/update-spec.sh              # download, extract, verify
#   scripts/update-spec.sh --url=URL    # override the download URL
#
# After running, manually update the constants in
# internal/cmd/gen-asc/spec/version.go so SpecSHA256 matches what the
# script prints.
set -euo pipefail

DEFAULT_URL="https://developer.apple.com/sample-code/app-store-connect/app-store-connect-openapi-specification.zip"
URL="${1:-$DEFAULT_URL}"
URL="${URL#--url=}"

DEST_DIR="internal/cmd/gen-asc/spec"
DEST_FILE="$DEST_DIR/app_store_connect_api_openapi.json"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading: $URL"
curl -sSL -o "$TMPDIR/download" "$URL"

# Detect payload: raw JSON, or ZIP containing a JSON file.
FIRST_BYTES="$(head -c 4 "$TMPDIR/download" | od -An -c | tr -d ' ')"

if [[ "$FIRST_BYTES" == PK* ]]; then
  echo "Payload is a ZIP; extracting..."
  mkdir -p "$TMPDIR/extract"
  unzip -q "$TMPDIR/download" -d "$TMPDIR/extract"
  # Pick the LARGEST .json file outside __MACOSX/. macOS-zipped archives
  # contain a tiny resource-fork sidecar (e.g. __MACOSX/._spec.json) that
  # would otherwise be selected first by find -name '*.json'.
  JSON_PATH="$(find "$TMPDIR/extract" -type f -name '*.json' \
                ! -path '*/__MACOSX/*' \
                -exec wc -c {} + 2>/dev/null \
                | sort -nr | awk 'NR==1 && $1>0 {print $2}')"
  if [[ -z "$JSON_PATH" || ! -f "$JSON_PATH" ]]; then
    echo "ERROR: no usable .json file inside the ZIP" >&2
    echo "ZIP contents:" >&2
    unzip -l "$TMPDIR/download" >&2
    exit 1
  fi
  echo "Selected: $JSON_PATH"
elif [[ "${FIRST_BYTES:0:1}" == "{" ]]; then
  echo "Payload is raw JSON."
  JSON_PATH="$TMPDIR/download"
else
  echo "ERROR: unknown payload format (starts with $FIRST_BYTES)" >&2
  echo "First 200 bytes:" >&2
  head -c 200 "$TMPDIR/download" >&2
  exit 1
fi

SHA="$(shasum -a 256 "$JSON_PATH" | awk '{print $1}')"
BYTES="$(wc -c < "$JSON_PATH" | tr -d ' ')"

mkdir -p "$DEST_DIR"
cp "$JSON_PATH" "$DEST_FILE"

echo
echo "Vendor file:  $DEST_FILE"
echo "Size (bytes): $BYTES"
echo "SHA256:       $SHA"
echo
echo "Next: paste the SHA256 into internal/cmd/gen-asc/spec/version.go"

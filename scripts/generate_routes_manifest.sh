#!/bin/sh
set -e
export PATH=/usr/local/go/bin:$PATH
mkdir -p runtime 2>/dev/null || true

# Run generator; allow it to write directly. Capture stdout to temp in case fallback JSON emitted.
TMP=runtime/.routes_manifest_out
go run ./cmd/routes-manifest > "$TMP" 2>&1 || true

if [ -f runtime/routes-manifest.json ]; then
  rm -f "$TMP"
  exit 0
fi

# Fallback: extract first JSON object from captured output
if grep -q 'generatedAt' "$TMP"; then
  awk 'BEGIN{p=0} /^{/{p=1} {if(p) print} /^}/{if(p){exit}}' "$TMP" > runtime/routes-manifest.json || true
fi
rm -f "$TMP"
[ -f runtime/routes-manifest.json ] || { echo "Failed to produce routes-manifest.json" >&2; exit 1; }
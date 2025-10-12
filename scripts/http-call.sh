#!/bin/bash
set -euo pipefail

# Load .env to infer BACKEND_URL if present
if [[ -f ".env" ]]; then
  while IFS='=' read -r key value; do
    [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
    value=$(echo "$value" | sed 's/^"\(.*\)"$/\1/' | sed "s/^'\(.*\)'$/\1/")
    export "$key=$value"
  done < .env
fi

# Prefer environment variables, fall back to positional args
METHOD="${METHOD:-${1:-GET}}"
ENDPOINT="${ENDPOINT:-${2:-/}}"
BODY="${BODY:-${3:-}}"
CONTENT_TYPE="${CONTENT_TYPE:-${4:-text/html}}"

if [[ -z "${BACKEND_URL:-}" ]]; then
  for host in gotrs-backend backend gotrs-ce-backend-1; do
    if getent hosts "$host" >/dev/null 2>&1; then
      BACKEND_URL="http://$host:8080"; break
    fi
  done
  BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
fi

# Build curl args
args=(-k -i -s -X "$METHOD" "$BACKEND_URL$ENDPOINT" -H "Accept: $CONTENT_TYPE")
if [[ -n "${AUTH_TOKEN:-}" ]]; then
  args+=(-H "Authorization: Bearer $AUTH_TOKEN")
fi
if [[ -n "$BODY" ]]; then
  args+=(-H "Content-Type: $CONTENT_TYPE" -d "$BODY")
fi

curl "${args[@]}" | sed -n '1,200p'

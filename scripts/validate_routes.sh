#!/bin/sh
set -eu

# Validate that NO static routes exist in htmx_routes.go.
# ALL routes must be defined in YAML files (routes/*.yaml).

SRC_FILE=internal/api/htmx_routes.go

# Find hardcoded route patterns like .GET("/path", .POST("/admin/users"
# Excludes generic helper functions that use variables (path, handlers...)
ROUTES=$(grep -E '\.(GET|POST|PUT|DELETE|PATCH)\s*\(\s*"/' "$SRC_FILE" 2>/dev/null | grep -v '^\s*//' || true)

if [ -z "$ROUTES" ]; then
  echo "✓ No static routes in htmx_routes.go. All routes defined in YAML."
  exit 0
fi

COUNT=$(echo "$ROUTES" | wc -l | tr -d ' ')
echo "ERROR: Found $COUNT hard-coded route(s) in htmx_routes.go:" >&2
echo "$ROUTES" >&2
echo "" >&2
echo "ALL routes must be defined in routes/*.yaml files." >&2
echo "htmx_routes.go should only contain initialization logic." >&2
exit 1

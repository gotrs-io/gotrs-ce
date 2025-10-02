#!/usr/bin/env bash
set -euo pipefail

# Ephemeral Go security scan script (container-first friendly)
# Assumes running inside a golang:<version>-alpine or debian image with bash and curl available.

echo "[ci-go-security] Starting Go security scan..."

ART_DIR="security-artifacts"
mkdir -p "${ART_DIR}"

# Ensure tools
export GOFLAGS="-buildvcs=false ${GOFLAGS:-}"

# Tidy & download (module graph consistent)
go mod tidy
go mod download

echo "[ci-go-security] Running govulncheck"
go install golang.org/x/vuln/cmd/govulncheck@latest
# Text output
if ! govulncheck ./... | tee "${ART_DIR}/govulncheck.txt"; then
  echo "[ci-go-security] govulncheck reported vulnerabilities (continuing)."
fi
# JSON (best effort; older versions may differ)
if govulncheck -json ./... > "${ART_DIR}/govulncheck.json" 2>/dev/null; then
  echo "[ci-go-security] govulncheck JSON written";
else
  echo "[ci-go-security] govulncheck JSON unavailable (ignored)";
fi

echo "[ci-go-security] Running gosec"
go install github.com/securego/gosec/v2/cmd/gosec@v2.21.0
# JSON output (for artifact)
gosec -conf .gosec.json -fmt json -out "${ART_DIR}/gosec-results.json" ./... || true
# Text output (for logs + artifact)
gosec -conf .gosec.json -fmt text ./... | tee "${ART_DIR}/gosec.txt" || true

echo "[ci-go-security] Running go vet"
if ! go vet ./...; then
  echo "[ci-go-security] go vet found issues" >&2
fi

echo "[ci-go-security] Installing golangci-lint"
GOLANGCI_LINT_VERSION=v1.55.2
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin ${GOLANGCI_LINT_VERSION}

echo "[ci-go-security] Running golangci-lint"
if ! golangci-lint run --timeout=5m -out-format json > "${ART_DIR}/golangci-lint.json"; then
  echo "[ci-go-security] golangci-lint issues detected" >&2
fi
golangci-lint run --timeout=5m > "${ART_DIR}/golangci-lint.txt" || true

echo "[ci-go-security] Artifact directory contents:"
ls -1 "${ART_DIR}" || true

echo "[ci-go-security] Completed Go security scan."

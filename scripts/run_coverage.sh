#!/usr/bin/env bash
set -euo pipefail

mkdir -p generated

PKGS=$(go list ./... | grep -Ev '/tests/|/tools/test-utilities|/examples$|/internal/api($|/)|/internal/i18n$')
if [[ -z "${PKGS}" ]]; then
	echo "No packages selected for coverage" >&2
	exit 1
fi

go test -buildvcs=false -v -race -coverprofile=generated/coverage.out -covermode=atomic ${PKGS}

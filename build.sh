#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
OUT="bin/todoapp"
[ "${GOOS:-$(go env GOOS)}" = "windows" ] && OUT="bin/todoapp.exe"

CGO_ENABLED=0 go build -v \
  -ldflags "-s -w -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
  -o "${OUT}" ./cmd/todoapp

echo "built ${OUT} (commit ${GIT_COMMIT}, ${BUILD_TIME})"

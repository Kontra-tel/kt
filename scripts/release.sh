#!/usr/bin/env bash
set -euo pipefail
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
TAG="$(git describe --tags --exact-match HEAD 2>/dev/null || true)"
VERSION="${VERSION:-${TAG#v}}"
if [ -z "$VERSION" ] || [ "$VERSION" = "$TAG" ]; then
  VERSION="0.0.0-dev.${COMMIT}"
fi
mkdir -p dist
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
  GOOS="${target%/*}"
  GOARCH="${target#*/}"
  OUT="dist/kt-${GOOS}-${GOARCH}"
  echo "building $OUT"
  GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o "$OUT" ./cmd/kt
done
(cd dist && sha256sum kt-* > SHA256SUMS)

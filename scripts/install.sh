#!/usr/bin/env bash
set -euo pipefail

API="https://git.kontra.tel/api/v1/repos/kontra.tel/Kt"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
ASSET="kt-${OS}-${ARCH}"

echo "Fetching latest kt release..."
RELEASE=$(curl -sf "${API}/releases/latest")
TAG=$(echo "$RELEASE" | grep '"tag_name"' | head -1 | cut -d'"' -f4)

URL=$(echo "$RELEASE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
want = sys.argv[1]
for a in data.get('assets', []):
    if a['name'] == want:
        print(a['browser_download_url'])
        break
else:
    print('', end='')
" "$ASSET")

if [ -z "$URL" ]; then
    echo "No binary found for ${OS}/${ARCH} in release ${TAG}" >&2
    exit 1
fi

echo "Installing kt ${TAG} (${OS}/${ARCH}) to ${BIN_DIR}/kt..."
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT
curl -sL "$URL" -o "$TMP"
chmod +x "$TMP"

if [ -w "$BIN_DIR" ]; then
    mv "$TMP" "${BIN_DIR}/kt"
else
    sudo mv "$TMP" "${BIN_DIR}/kt"
fi

echo "Done."
kt version

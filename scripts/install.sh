#!/usr/bin/env bash
set -euo pipefail

API="https://git.kontra.tel/api/v1/repos/kontra.tel/Kt"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
KT_VERSION="${KT_VERSION:-}"
KT_PRERELEASE="${KT_PRERELEASE:-0}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
ASSET="kt-${OS}-${ARCH}"

if [ -n "$KT_VERSION" ]; then
    case "$KT_VERSION" in
        v*) TAG="$KT_VERSION" ;;
        *) TAG="v$KT_VERSION" ;;
    esac
    echo "Fetching kt release ${TAG}..."
    RELEASE=$(curl -sf "${API}/releases/tags/${TAG}")
elif [ "$KT_PRERELEASE" = "1" ] || [ "$KT_PRERELEASE" = "true" ]; then
    echo "Fetching latest kt prerelease..."
    RELEASE=$(curl -sf "${API}/releases" | python3 -c "
import json, sys

def parse_version(tag):
    tag = tag.lstrip('v')
    main, *rest = tag.split('-', 1)
    major, minor, patch = [int(x) for x in main.split('.')]
    if not rest:
        return (major, minor, patch, 1, 999, '')
    label, num = rest[0].rsplit('.', 1)
    rank = {'alpha': 0, 'beta': 1, 'rc': 2}.get(label, 3)
    return (major, minor, patch, 0, rank, int(num))

releases = json.load(sys.stdin)
prereleases = [r for r in releases if r.get('prerelease')]
if not prereleases:
    raise SystemExit('No prereleases found')
print(json.dumps(max(prereleases, key=lambda r: parse_version(r['tag_name']))))
")
    TAG=$(echo "$RELEASE" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
else
    echo "Fetching latest kt release..."
    RELEASE=$(curl -sf "${API}/releases/latest")
    TAG=$(echo "$RELEASE" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
fi

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

SUMS_URL=$(echo "$RELEASE" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for a in data.get('assets', []):
    if a['name'] == 'SHA256SUMS':
        print(a['browser_download_url'])
        break
else:
    print('', end='')
")

if [ -z "$URL" ]; then
    echo "No binary found for ${OS}/${ARCH} in release ${TAG}" >&2
    exit 1
fi
if [ -z "$SUMS_URL" ]; then
    echo "No SHA256SUMS found in release ${TAG}" >&2
    exit 1
fi

echo "Installing kt ${TAG} (${OS}/${ARCH}) to ${BIN_DIR}/kt..."
TMP=$(mktemp)
SUMS=$(mktemp)
trap 'rm -f "$TMP" "$SUMS"' EXIT
curl -sL "$URL" -o "$TMP"
curl -sL "$SUMS_URL" -o "$SUMS"
python3 -c "
import hashlib, pathlib, sys
asset, binary, sums = sys.argv[1:]
want = None
for line in pathlib.Path(sums).read_text().splitlines():
    parts = line.split()
    if len(parts) >= 2 and parts[1] == asset:
        want = parts[0].lower()
        break
if not want:
    raise SystemExit(f'No checksum for {asset}')
got = hashlib.sha256(pathlib.Path(binary).read_bytes()).hexdigest()
if got != want:
    raise SystemExit(f'Checksum mismatch for {asset}')
" "$ASSET" "$TMP" "$SUMS"
chmod +x "$TMP"

if [ -w "$BIN_DIR" ]; then
    mv "$TMP" "${BIN_DIR}/kt"
else
    sudo mv "$TMP" "${BIN_DIR}/kt"
fi

echo "Done."
kt version

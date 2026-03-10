#!/usr/bin/env bash
set -euo pipefail

VERSION="${VERSION:-dev}"
LDFLAGS="-s -w -X main.Version=${VERSION}"
MODULE="github.com/openclaw/openclaw-installer"
DIST="dist"

echo "Building openclaw-installer ${VERSION}"

# Clean
rm -rf "${DIST}"
mkdir -p "${DIST}"

# Build targets: OS/ARCH
TARGETS=(
  "windows/amd64/.exe"
  "darwin/amd64/"
  "darwin/arm64/"
)

for target in "${TARGETS[@]}"; do
  IFS='/' read -r goos goarch ext <<< "${target}"
  output="${DIST}/openclaw-installer-${goos}-${goarch}${ext}"
  echo "  -> ${output}"
  CGO_ENABLED=0 GOOS="${goos}" GOARCH="${goarch}" \
    go build -ldflags="${LDFLAGS}" -o "${output}" .
done

# Create macOS universal binary (only possible on macOS with lipo)
if [[ "$(uname -s)" == "Darwin" ]] && command -v lipo &>/dev/null; then
  echo "  -> ${DIST}/openclaw-installer-darwin-universal (lipo)"
  lipo -create \
    "${DIST}/openclaw-installer-darwin-amd64" \
    "${DIST}/openclaw-installer-darwin-arm64" \
    -output "${DIST}/openclaw-installer-darwin-universal"
fi

# Generate checksums
echo "Generating checksums..."
cd "${DIST}"
if command -v sha256sum &>/dev/null; then
  sha256sum * > checksums.txt
elif command -v shasum &>/dev/null; then
  shasum -a 256 * > checksums.txt
else
  echo "WARNING: no sha256sum or shasum found, skipping checksums" >&2
fi
cd ..

echo "Done. Artifacts in ${DIST}/"
ls -lh "${DIST}/"

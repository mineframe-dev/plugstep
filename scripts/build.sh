#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-dev}"
DIST_DIR="./dist"
MODULE="forgejo.perny.dev/mineframe/plugstep/cmd/plugstep"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

echo "Building plugstep $VERSION"

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"

    output_name="plugstep-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    echo "  Building $GOOS/$GOARCH..."
    GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build \
        -ldflags="-s -w -X main.Version=$VERSION" \
        -o "$DIST_DIR/$output_name" \
        "$MODULE"
done

echo "Generating checksums..."
(cd "$DIST_DIR" && sha256sum plugstep-* > checksums.txt)

echo "Done. Binaries in $DIST_DIR/"
ls -la "$DIST_DIR"

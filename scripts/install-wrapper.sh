#!/usr/bin/env bash
set -euo pipefail

WRAPPER_BASE_URL="${WRAPPER_BASE_URL:-https://releases.perny.dev/mineframe/plugstep-wrapper}"
RELEASES_BASE_URL="${RELEASES_BASE_URL:-https://releases.perny.dev/mineframe/plugstep}"

echo "Installing plugstepw..."

download() {
    local url="$1"
    local output="$2"
    if command -v curl &> /dev/null; then
        curl -fSL "$url" -o "$output"
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$output"
    else
        echo "Error: curl or wget required" >&2
        exit 1
    fi
}

# Download bash wrapper
download "$WRAPPER_BASE_URL/plugstepw" plugstepw
chmod +x plugstepw

# Download PowerShell wrapper
download "$WRAPPER_BASE_URL/plugstepw.ps1" plugstepw.ps1

# Fetch and set latest version
echo "Fetching latest version..."
download "$RELEASES_BASE_URL/latest" .plugstep-version

VERSION=$(cat .plugstep-version)
echo "Installed plugstepw (bash + PowerShell)"
echo "Set version to $VERSION"
echo "Run ./plugstepw to get started"

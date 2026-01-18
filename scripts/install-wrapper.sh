#!/usr/bin/env bash
set -euo pipefail

WRAPPER_BASE_URL="${WRAPPER_BASE_URL:-https://releases.perny.dev/mineframe/plugstep-wrapper}"

echo "Installing plugstepw..."

if command -v curl &> /dev/null; then
    curl -fSL "$WRAPPER_BASE_URL/plugstepw" -o plugstepw
elif command -v wget &> /dev/null; then
    wget -q "$WRAPPER_BASE_URL/plugstepw" -O plugstepw
else
    echo "Error: curl or wget required" >&2
    exit 1
fi

chmod +x plugstepw

echo "Installed plugstepw to current directory."
echo "Create a .plugstep-version file with your desired version (e.g., v1.0.0)"

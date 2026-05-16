#!/usr/bin/env bash
#
# Build h2.core as iOS Framework and Android AAR using gomobile.
#
# Prerequisites:
#   - Go 1.21+
#   - Xcode (for iOS)
#   - Android SDK/NDK (for Android)
#   - gomobile: go install golang.org/x/mobile/cmd/gomobile@latest
#
# Usage:
#   ./build/mobile.sh              # Build both iOS and Android
#   ./build/mobile.sh ios          # Build iOS only
#   ./build/mobile.sh android      # Build Android only
#
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

OUT_DIR="${OUT_DIR:-dist/mobile}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")}"

mkdir -p "$OUT_DIR"

# Add Go bin to PATH
export PATH="$PATH:$(go env GOPATH)/bin"

# Check gomobile
check_gomobile() {
    if ! command -v gomobile &> /dev/null; then
        echo "gomobile not found. Installing..."
        go install golang.org/x/mobile/cmd/gomobile@latest
        echo "Initializing gomobile..."
        gomobile init
    fi
}

# Build iOS Framework
build_ios() {
    echo "=== Building iOS Framework ==="
    echo "Output: $OUT_DIR/H2Core.xcframework"
    echo ""

    # Check for Xcode
    if ! command -v xcodebuild &> /dev/null; then
        echo "WARNING: Xcode not found. Skipping iOS build."
        return 1
    fi

    gomobile bind \
        -target=ios \
        -ldflags="-s -w -X github.com/vpnclient/https-vpn/mobile.Version=$VERSION" \
        -o "$OUT_DIR/H2Core.xcframework" \
        ./mobile

    echo "iOS Framework built successfully!"
    echo ""
}

# Build Android AAR
build_android() {
    echo "=== Building Android AAR ==="
    echo "Output: $OUT_DIR/h2core.aar"
    echo ""

    # Check for Android SDK
    if [[ -z "${ANDROID_HOME:-}" ]] && [[ -z "${ANDROID_SDK_ROOT:-}" ]]; then
        echo "WARNING: ANDROID_HOME/ANDROID_SDK_ROOT not set."
        echo "Attempting build anyway (may fail if SDK not in default location)..."
    fi

    gomobile bind \
        -target=android \
        -androidapi=21 \
        -ldflags="-s -w -X github.com/vpnclient/https-vpn/mobile.Version=$VERSION" \
        -o "$OUT_DIR/h2core.aar" \
        ./mobile

    echo "Android AAR built successfully!"
    echo ""
}

# Main
check_gomobile

TARGET="${1:-all}"

case "$TARGET" in
    ios)
        build_ios
        ;;
    android)
        build_android
        ;;
    all)
        # Try both, continue if one fails
        build_ios || echo "iOS build skipped/failed"
        build_android || echo "Android build skipped/failed"
        ;;
    *)
        echo "Usage: $0 [ios|android|all]"
        exit 1
        ;;
esac

echo "=== Build Complete ==="
echo "Outputs in $OUT_DIR/:"
ls -la "$OUT_DIR"/ 2>/dev/null || true

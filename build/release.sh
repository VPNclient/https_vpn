#!/bin/bash
# h2.core cross-platform release build script
# Builds CLI binaries for all supported platforms

set -e

VERSION="${VERSION:-0.1.0}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"
CMD_PATH="./cmd/https-vpn"
BINARY_NAME="https-vpn"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build targets: GOOS/GOARCH/ARM/Name
TARGETS=(
    # Linux
    "linux/amd64//h2-linux-64"
    "linux/386//h2-linux-32"
    "linux/arm64//h2-linux-arm64-v8a"
    "linux/arm/7/h2-linux-arm32-v7a"
    "linux/arm/6/h2-linux-arm32-v6"
    "linux/arm/5/h2-linux-arm32-v5"
    "linux/mips//h2-linux-mips32"
    "linux/mipsle//h2-linux-mips32le"
    "linux/mips64//h2-linux-mips64"
    "linux/mips64le//h2-linux-mips64le"
    "linux/ppc64//h2-linux-ppc64"
    "linux/ppc64le//h2-linux-ppc64le"
    "linux/riscv64//h2-linux-riscv64"
    "linux/s390x//h2-linux-s390x"
    "linux/loong64//h2-linux-loong64"

    # macOS
    "darwin/amd64//h2-macos-64"
    "darwin/arm64//h2-macos-arm64-v8a"

    # FreeBSD
    "freebsd/amd64//h2-freebsd-64"
    "freebsd/386//h2-freebsd-32"
    "freebsd/arm64//h2-freebsd-arm64-v8a"
    "freebsd/arm/7/h2-freebsd-arm32-v7a"

    # OpenBSD
    "openbsd/amd64//h2-openbsd-64"
    "openbsd/386//h2-openbsd-32"
    "openbsd/arm64//h2-openbsd-arm64-v8a"
    "openbsd/arm/7/h2-openbsd-arm32-v7a"

    # Android (CLI)
    "android/amd64//h2-android-amd64"
    "android/arm64//h2-android-arm64-v8a"

    # Windows
    "windows/amd64//h2-windows-64"
    "windows/386//h2-windows-32"
    "windows/arm64//h2-windows-arm64-v8a"
)

echo "Building h2.core v${VERSION} for ${#TARGETS[@]} platforms..."

build_target() {
    local target="$1"
    IFS='/' read -r goos goarch goarm name <<< "$target"

    local binary="$BINARY_NAME"
    [[ "$goos" == "windows" ]] && binary="${BINARY_NAME}.exe"

    local build_dir="$OUTPUT_DIR/$name"
    mkdir -p "$build_dir"

    echo "Building $name..."

    local env="GOOS=$goos GOARCH=$goarch CGO_ENABLED=0"
    [[ -n "$goarm" ]] && env="$env GOARM=$goarm"

    if eval $env go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" \
        -o "$build_dir/$binary" "$CMD_PATH" 2>/dev/null; then
        # Create zip
        (cd "$OUTPUT_DIR" && zip -q -r "$name.zip" "$name")
        rm -rf "$build_dir"
        echo "  -> $OUTPUT_DIR/$name.zip"
    else
        echo "  -> SKIPPED (build failed)"
        rm -rf "$build_dir"
    fi
}

# Build all targets
for target in "${TARGETS[@]}"; do
    build_target "$target"
done

# Win7 builds (requires Go 1.20 - last version with Win7 support)
# Skip if GO120 not set
if [ -n "$GO120" ] && [ -x "$GO120" ]; then
    echo ""
    echo "Building Win7 targets with Go 1.20..."

    for arch in "amd64/h2-win7-64" "386/h2-win7-32"; do
        IFS='/' read -r goarch name <<< "$arch"
        mkdir -p "$OUTPUT_DIR/$name"
        echo "Building $name..."
        GOOS=windows GOARCH=$goarch CGO_ENABLED=0 $GO120 build -trimpath \
            -ldflags="-s -w -X main.Version=${VERSION}" \
            -o "$OUTPUT_DIR/$name/${BINARY_NAME}.exe" "$CMD_PATH"
        (cd "$OUTPUT_DIR" && zip -q -r "$name.zip" "$name")
        rm -rf "$OUTPUT_DIR/$name"
        echo "  -> $OUTPUT_DIR/$name.zip"
    done
else
    echo ""
    echo "Skipping Win7 builds (set GO120=/path/to/go1.20/bin/go to enable)"
fi

# Generate checksums
echo ""
echo "Generating checksums..."
(cd "$OUTPUT_DIR" && sha256sum *.zip > checksums.txt 2>/dev/null || shasum -a 256 *.zip > checksums.txt)
echo "  -> $OUTPUT_DIR/checksums.txt"

echo ""
echo "Build complete! Archives in $OUTPUT_DIR/"
ls -lh "$OUTPUT_DIR"/*.zip 2>/dev/null | head -20
echo "..."
echo "Total: $(ls -1 "$OUTPUT_DIR"/*.zip 2>/dev/null | wc -l) archives"

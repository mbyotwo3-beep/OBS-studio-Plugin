#!/bin/bash
# Setup script for Breez SDK Spark (Linux/Mac)

set -e

echo "=== Breez SDK Spark Setup Script ==="
echo ""

# Check for Rust/Cargo
if ! command -v cargo &> /dev/null; then
    echo "Error: Rust/Cargo not found!"
    echo "Please install Rust from: https://rustup.rs/"
    exit 1
fi

echo "✓ Rust/Cargo found"
echo ""

# Navigate to SDK directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$SCRIPT_DIR/../third_party/breez_sdk"

if [ ! -d "$SDK_DIR" ]; then
    echo "Error: Breez SDK not found at $SDK_DIR"
    echo "Please run: git clone https://github.com/breez/spark-sdk.git third_party/breez_sdk"
    exit 1
fi

cd "$SDK_DIR"
echo "Building Breez SDK Spark..."
echo "This may take several minutes..."
echo ""

# Build the SDK with release optimizations
cargo build --release --package breez-sdk-core

echo ""
echo "=== Build Complete ==="
echo ""
echo "Library location: $SDK_DIR/target/release/"
echo ""
echo "Next steps:"
echo "1. Configure CMake with: cmake -B build -S ."
echo "2. Build the OBS plugin: cmake --build build --config Release"
echo ""

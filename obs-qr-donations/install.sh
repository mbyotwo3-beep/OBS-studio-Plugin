#!/bin/bash
# OBS QR Donations Plugin - Easy Installer (Linux/Mac)

echo "========================================"
echo " OBS QR Donations Plugin - Installer"
echo "========================================"
echo ""

# Find OBS plugin directory
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    PLUGIN_DIR="$HOME/.config/obs-studio/plugins/obs-qr-donations/bin/64bit"
    DATA_DIR="$HOME/.config/obs-studio/plugins/obs-qr-donations/data"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    PLUGIN_DIR="$HOME/Library/Application Support/obs-studio/plugins/obs-qr-donations/bin"
    DATA_DIR="$HOME/Library/Application Support/obs-studio/plugins/obs-qr-donations/data"
else
    echo "[ERROR] Unsupported operating system"
    exit 1
fi

echo "[INFO] Installing to: $PLUGIN_DIR"
echo ""

# Create directories
mkdir -p "$PLUGIN_DIR"
mkdir -p "$DATA_DIR"

# Check if plugin library exists
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    PLUGIN_FILE="build/libobs-qr-donations.so"
else
    PLUGIN_FILE="build/libobs-qr-donations.dylib"
fi

if [ ! -f "$PLUGIN_FILE" ]; then
    echo "[WARNING] Plugin not built yet!"
    echo ""
    echo "Please build the plugin first:"
    echo "  1. Install dependencies (see README.md)"
    echo "  2. Run: cmake -B build -S . -DBREEZ_USE_STUB=ON"
    echo "  3. Run: cmake --build build --config Release"
    echo "  4. Run this installer again"
    echo ""
    exit 1
fi

# Copy plugin files
echo "Installing plugin files..."
cp "$PLUGIN_FILE" "$PLUGIN_DIR/" || { echo "[ERROR] Failed to copy plugin"; exit 1; }

# Copy data files if they exist
if [ -d "data" ]; then
    cp -r data/* "$DATA_DIR/"
fi

echo ""
echo "========================================"
echo " Installation Complete!"
echo "========================================"
echo ""
echo "Plugin installed to: $PLUGIN_DIR"
echo "Data files at: $DATA_DIR"
echo ""
echo "To use the plugin:"
echo "  1. Open OBS Studio"
echo "  2. Click + in Sources panel"
echo "  3. Select 'QR Donations'"
echo "  4. Configure your cryptocurrency addresses"
echo ""
echo "For Lightning Network support:"
echo "  - Get a free Breez API key from https://breez.technology/"
echo "  - Enter it in the plugin settings"
echo ""

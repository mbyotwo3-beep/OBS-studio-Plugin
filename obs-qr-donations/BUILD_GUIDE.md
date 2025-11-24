# Building OBS QR Donations Plugin

## Prerequisites

Before building, you need:

### Required
- **CMake** 3.16+ ([Download](https://cmake.org/download/))
- **Git** ([Download](https://git-scm.com/))
- **C++ Compiler**:
  - Windows: Visual Studio 2019+ ([Download](https://visualstudio.microsoft.com/downloads/))
  - Linux: GCC 9+ or Clang 10+
  - Mac: Xcode Command Line Tools

### OBS Studio Development Environment

You have two options:

#### Option 1: Use Pre-built OBS (Easier)
1. Download OBS Studio from https://obsproject.com/
2. Download OBS Studio development files:
   - Windows: https://github.com/obsproject/obs-studio/releases (look for `-windows-x64.zip`)
   - Linux: Install `obs-studio-dev` package
   - Mac: Build from source or use Homebrew

#### Option 2: Build OBS from Source (Advanced)
```bash
git clone --recursive https://github.com/obsproject/obs-studio.git
cd obs-studio
cmake -B build -S .
cmake --build build
```

### Qt 6.x
- **Windows**: Download from https://www.qt.io/download-qt-installer
- **Linux**: `sudo apt-get install qt6-base-dev` (Ubuntu/Debian)
- **Mac**: `brew install qt@6`

### Additional Libraries
- **qrencode**: QR code generation
  - Windows: `vcpkg install qrencode:x64-windows`
  - Linux: `sudo apt-get install libqrencode-dev`
  -Mac: `brew install qrencode`

## Quick Build (Stub Mode)

This builds without Breez SDK (no real Lightning payments, but works for testing):

### Windows
```powershell
# Configure
cmake -B build -S . -DBREEZ_USE_STUB=ON -DLibOBS_DIR="C:/path/to/obs-studio/build/libobs" -DQt6_DIR="C:/Qt/6.x.x/msvc2019_64/lib/cmake/Qt6"

# Build
cmake --build build --config Release

# Install
install.bat
```

### Linux
```bash
# Configure
cmake -B build -S . -DBREEZ_USE_STUB=ON

# Build
cmake --build build --config Release

# Install
chmod +x install.sh
./install.sh
```

### Mac
```bash
# Configure  
cmake -B build -S . -DBREEZ_USE_STUB=ON

# Build
cmake --build build --config Release

# Install
chmod +x install.sh
./install.sh
```

## Full Build (With Lightning Support)

### Step 1: Install Rust
Lightning support requires the Breez SDK, which needs Rust:
```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Windows: Download from https://rustup.rs/
```

### Step 2: Build Breez SDK
```bash
# Windows
scripts\setup_spark_sdk.bat

# Linux/Mac
chmod +x scripts/setup_spark_sdk.sh
./scripts/setup_spark_sdk.sh
```

This takes 5-15 minutes depending on your system.

### Step 3: Build Plugin
```bash
# Configure (without -DBREEZ_USE_STUB flag)
cmake -B build -S . -DLibOBS_DIR="path/to/obs/build/libobs"

# Build
cmake --build build --config Release

# Install
install.bat    # Windows
./install.sh   # Linux/Mac
```

## Troubleshooting Build Issues

### CMake Can't Find OBS
```
CMake Error: Could not find a package configuration file provided by "LibOBS"
```

**Solution**: Specify OBS location:
```bash
cmake -B build -S . -DLibOBS_DIR="C:/path/to/obs/build/libobs"
```

### CMake Can't Find Qt
```
CMake Error: Could not find a package configuration file provided by "Qt6"
```

**Solution**: Specify Qt location:
```bash
cmake -B build -S . -DQt6_DIR="C:/Qt/6.5.0/msvc2019_64/lib/cmake/Qt6"
```

### qrencode Not Found
```
CMake Error: Could not find qrencode
```

**Solution**:
- Windows: Use vcpkg: `vcpkg install qrencode:x64-windows`
- Linux: `sudo apt-get install libqrencode-dev`
- Mac: `brew install qrencode`

### Breez SDK Build Fails
```
cargo build failed
```

**Solutions**:
1. Make sure Rust is installed: `cargo --version`
2. Update Rust: `rustup update`
3. Check internet connection (downloads dependencies)
4. Try building manually:
   ```bash
   cd third_party/breez_sdk
   cargo build --release --package breez-sdk-core
   ```

### Visual Studio Not Found (Windows)
```
error: could not find instance of Visual Studio
```

**Solution**:
1. Install Visual Studio 2019 or later
2. Include "Desktop development with C++" workload
3. Restart command prompt after installation

## Verifying the Build

After building, you should have:
- Windows: `build/Release/obs-qr-donations.dll`
- Linux: `build/libobs-qr-donations.so`
- Mac: `build/libobs-qr-donations.dylib`

Verify the file exists:
```bash
# Windows
dir build\Release\obs-qr-donations.dll

# Linux/Mac
ls -lh build/libobs-qr-donations.*
```

## Installing to OBS

###Automatic (Recommended)
```bash
# Windows
install.bat

# Linux/Mac
chmod +x install.sh
./install.sh
```

### Manual Installation

#### Windows
Copy `build/Release/obs-qr-donations.dll` to:
- `C:\Program Files\obs-studio\obs-plugins\64bit\`

#### Linux
Copy `build/libobs-qr-donations.so` to:
- `~/.config/obs-studio/plugins/obs-qr-donations/bin/64bit/`

#### Mac
Copy `build/libobs-qr-donations.dylib` to:
- `~/Library/Application Support/obs-studio/plugins/obs-qr-donations/bin/`

## Testing

1. Start OBS Studio
2. Check Help → Log Files → View Current Log
3. Look for "QR Donations" plugin loaded message
4. In Sources, click + → should see "QR Donations" option
5. Add source and configure addresses
6. Verify QR code displays correctly

## Next Steps

- See [USER_GUIDE.md](USER_GUIDE.md) for usage instructions
- See [README.md](README.md) for feature overview
- Get Breez API key from https://breez.technology/ for Lightning support

## Common CMake Options

```bash
cmake -B build -S . \
  -DBREEZ_USE_STUB=ON \              # Use stub (no real payments)
  -DBREEZ_STUB_SIMULATE=ON \          # Simulate payments in stub mode
  -DCMAKE_BUILD_TYPE=Release \        # Release build (optimized)
  -DLibOBS_DIR=/path/to/obs \         # OBS SDK location
  -DQt6_DIR=/path/to/qt/cmake/Qt6     # Qt6 location
```

## Getting Help

If you encounter issues:
1. Check this guide's **Troubleshooting** section
2. Review CMake error messages carefully
3. Check that all prerequisites are installed
4. Try building in stub mode first
5. Check OBS plugin development docs: https://obsproject.com/docs/plugins/

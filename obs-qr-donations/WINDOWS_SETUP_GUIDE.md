# Windows Build Setup Guide

## Current Status

✅ **Prerequisites Installed:**
- CMake 4.2.0
- Git 2.50.1

❌ **Missing:**
- OBS Studio SDK (required for plugin development)
- Qt6 (possibly)

## Quick Setup Steps

### Step 1: Install OBS Studio (if not already installed)

Download and install OBS Studio from:
https://obsproject.com/download

### Step 2: Get OBS Studio SDK

Option A: Download pre-built SDK (Easiest)
1. Go to: https://github.com/obsproject/obs-studio/releases
2. Download the latest `OBS-Studio-XX.X.X-windows-x64.zip`
3. Extract to a known location (e.g., `C:\OBS-SDK`)

Option B: Build from source (Advanced)
```powershell
git clone --recursive https://github.com/obsproject/obs-studio.git C:\obs-studio
cd C:\obs-studio
cmake -B build -S . -G "Visual Studio 17 2022"
cmake --build build --config Release
```

### Step 3: Install Qt6

Download Qt6 from:
https://www.qt.io/download-qt-installer

Choose "Open Source" and install to: `C:\Qt`

### Step 4: Configure Plugin Build

Once OBS SDK and Qt6 are installed:

```powershell
cd "C:\Users\Administrator\Desktop\OBS studio Plugin\obs-qr-donations"

# Configure with paths
cmake -B build -S . `
  -DBREEZ_USE_STUB=ON `
  -DBREEZ_STUB_SIMULATE=ON `
  -DLibOBS_DIR="C:\OBS-SDK\cmake" `
  -DQt6_DIR="C:\Qt\6.x.x\msvc2019_64\lib\cmake\Qt6" `
  -G "Visual Studio 17 2022"

# Build
cmake --build build --config Release
```

### Step 5: Install Plugin

```powershell
.\install.bat
```

## Alternative: Use vcpkg for Dependencies

```powershell
# Install vcpkg
git clone https://github.com/Microsoft/vcpkg.git C:\vcpkg
cd C:\vcpkg
.\bootstrap-vcpkg.bat

# Install dependencies
.\vcpkg install qt6:x64-windows
.\vcpkg install qrencode:x64-windows

# Use vcpkg toolchain in cmake
cmake -B build -S . `
  -DBREEZ_USE_STUB=ON `
  -DCMAKE_TOOLCHAIN_FILE=C:\vcpkg\scripts\buildsystems\vcpkg.cmake `
  -G "Visual Studio 17 2022"
```

## Simplified Build (No OBS SDK Required)

If you just want to test the QR generation and plugin logic without OBS:

1. Use the test scripts:
```powershell
cd obs-qr-donations
python scripts/test_plugin.py  # GUI test tool
python scripts/run_integration_test.py  # Integration tests
```

2. Requirements:
```powershell
pip install qrcode pillow PyQt6
```

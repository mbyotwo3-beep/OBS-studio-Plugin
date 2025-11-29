# Build Troubleshooting Guide

This document covers common build issues and their solutions.

## Common Issues

### 1. CMake Not Found
**Error:** `'cmake' is not recognized as an internal or external command`

**Solution:**
- Install CMake from https://cmake.org/download/
- Or install via chocolatey: `choco install cmake`
- Add CMake to PATH

### 2. Qt6 Not Found
**Error:** `Could not find a package configuration file provided by "Qt6"`

**Solutions:**
```powershell
# Option 1: Set Qt6_DIR manually
cmake -B build -S . -DQt6_DIR="C:/Qt/6.5.3/msvc2019_64/lib/cmake/Qt6"

# Option 2: Install Qt6
# Download from https://www.qt.io/download-qt-installer
# Install MSVC 2019 64-bit component
```

### 3. OBS SDK Not Found
**Error:** `Could not find a package configuration file provided by "LibOBS"`

**Solutions:**
```powershell
# Option 1: Build OBS from source
git clone https://github.com/obsproject/obs-studio
cd obs-studio
# Follow OBS build instructions

# Option 2: Point to existing OBS installation
cmake -B build -S . -DLibOBS_DIR="C:/Program Files/obs-studio/cmake"
```

### 4. Missing Qt6::Multimedia
**Error:** `Qt6::Multimedia not found`

**Solution:**
- Reinstall Qt6 with Multimedia component
- Or install via Qt Maintenance Tool:
  - Add Components → Qt Multimedia

### 5. MSVC Compiler Not Found
**Error:** `No suitable compiler found`

**Solutions:**
```powershell
# Install Visual Studio 2019 or 2022
# Include "Desktop development with C++" workload

# Or use Visual Studio Developer Command Prompt
# Find it in Start Menu under Visual Studio folder
```

### 6. vcpkg Not Found
**Error:** `vcpkg/scripts/buildsystems/vcpkg.cmake not found`

**Solution:**
```powershell
# Clone vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
.\bootstrap-vcpkg.bat
.\vcpkg install qrencode:x64-windows
```

### 7. Build Errors in breez-service.cpp
**Error:** Compilation errors related to Breez SDK

**Solution:**
- The plugin defaults to stub mode (no Breez SDK required)
- Ensure `BREEZ_USE_STUB=ON` in cmake command:
  ```powershell
  cmake -B build -S . -DBREEZ_USE_STUB=ON
  ```

### 8. Missing QSoundEffect
**Error:** `QSoundEffect: No such file or directory`

**Solution:**
- Verify Qt6::Multimedia is in CMakeLists.txt
- Rebuild CMake configuration:
  ```powershell
  rmdir /s /q build
  cmake -B build -S .
  ```

## Verification Steps

After fixing issues, verify your build:

```powershell
# Run test script
powershell -ExecutionPolicy Bypass -File test-plugin.ps1

# Check DLL was created
dir build\Release\obs-qr-donations.dll

# Try installation
.\install.bat
```

## Clean Build

If all else fails, try a clean build:

```powershell
# Remove all build artifacts
rmdir /s /q build
rmdir /s /q CMakeCache.txt
rmdir /s /q CMakeFiles

# Reconfigure
cmake -B build -S . -DBREEZ_USE_STUB=ON -DQt6_DIR="YOUR_QT_PATH"

# Rebuild
cmake --build build --config Release
```

## Getting Help

If you're still stuck:

1. Check OBS logs: `Help → Log Files → View Current Log`
2. Look for `[QR Donations]` or `Breez` entries
3. Check the error messages carefully
4. Verify all prerequisites are installed:
   - CMake 3.16+
   - Qt 6.x
   - Visual Studio 2019/2022
   - vcpkg (optional, for qrencode)

## Build Success Checklist

✅ CMake found in PATH  
✅ Qt6 installed with MSVC component  
✅ Qt6::Multimedia component present  
✅ Visual Studio C++ compiler available  
✅ `build/Release/obs-qr-donations.dll` created  
✅ Test script passes

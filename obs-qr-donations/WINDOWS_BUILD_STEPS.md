# Step-by-Step Build Instructions for Windows

## What You Have ✅
- CMake 4.2.0
- Visual Studio 2019
- OBS Studio

## What We Need to Install

### Step 1: Install Qt6 (Required)

**Quick Install via Online Installer:**
1. Download Qt Online Installer: https://www.qt.io/download-qt-installer
2. Run installer
3. Select: Qt 6.5.3 (or latest 6.x)
4. Choose components:
   - MSVC 2019 64-bit ✓
   - Qt5 Compatibility Module ✓
5. Note installation path (e.g., `C:\Qt\6.5.3\msvc2019_64`)

### Step 2: Install vcpkg + qrencode

Open PowerShell in the project directory and run:

```powershell
# Clone vcpkg
git clone https://github.com/Microsoft/vcpkg.git
cd vcpkg
.\bootstrap-vcpkg.bat
.\vcpkg install qrencode:x64-windows
cd ..
```

### Step 3: Get OBS SDK

**Option A: Use Installed OBS** (Simpler)
Your OBS installation should work, but we may need headers.

**Option B: Download OBS SDK**
```powershell
# Download and extract OBS development files
# We'll do this if needed
```

## Build Commands

Once everything is installed, run these commands in PowerShell:

```powershell
# Configure (replace Qt path with your actual path)
cmake -B build -S . `
  -DBREEZ_USE_STUB=ON `
  -DCMAKE_TOOLCHAIN_FILE=vcpkg/scripts/buildsystems/vcpkg.cmake `
  -DQt6_DIR="C:/Qt/6.5.3/msvc2019_64/lib/cmake/Qt6"

# Build
cmake --build build --config Release

# Install to OBS
.\install.bat
```

## Let's Start!

**Ready to begin? Here's what we'll do:**
1. First, let me check if you have Git installed
2. Then we'll install vcpkg and qrencode (5 minutes)
3. You install Qt6 (10-15 minutes)
4. We build the plugin (2 minutes)
5. Test in OBS!

**Current Status:** Checking for Git...

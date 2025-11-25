# Quick Start - OBS QR Donations Plugin

## ⚡ Fastest Way to Build (Windows)

### Option 1: Automated vcpkg Build (Recommended)

This will automatically install all dependencies and build the plugin:

```powershell
cd "C:\Users\Administrator\Desktop\OBS studio Plugin\obs-qr-donations"
.\scripts\build-with-vcpkg.ps1
```

**What it does:**
- ✅ Installs vcpkg (package manager)
- ✅ Downloads Qt6 and qrencode automatically
- ✅ Builds the plugin
- ⏳ First run: ~30 minutes (installs dependencies)
- ⏳ Subsequent builds: ~2 minutes

### Option 2: Test Without Building

Test the plugin functionality without compiling:

```powershell
# Demo QR generation
python demo_qr_functionality.py

# GUI test tool  
pip install PyQt6
python scripts\test_plugin.py
```

## Current Status

✅ **Working:**
- QR generation (tested & verified)
- Python test suite
- Build automation scripts

⚠️ **For full OBS plugin:**
- Run the vcpkg build script above
- OR follow manual SDK setup in WINDOWS_SETUP_GUIDE.md

## Manual Build (If vcpkg fails)

See [WINDOWS_SETUP_GUIDE.md](WINDOWS_SETUP_GUIDE.md) for manual VCM SDK setup instructions.

## After Building

```powershell
# Install to OBS
.\install.bat

# Start OBS and test
# Sources → + → QR Donations
```

## Need Help?

- **Build issues:** Check [BUILD_GUIDE.md](BUILD_GUIDE.md)
- **Windows specific:** See [WINDOWS_SETUP_GUIDE.md](WINDOWS_SETUP_GUIDE.md)
- **Testing:** See [scripts/README.md](scripts/README.md)

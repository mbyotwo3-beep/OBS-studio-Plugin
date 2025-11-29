# Quick Start - Build & Release

## 🚀 Build the Plugin (1 command!)

```powershell
.\build-quick.bat
```

This script will:
- ✅ Auto-detect Qt6 installation
- ✅ Configure CMake
- ✅ Build the plugin
- ✅ Show clear error messages if something goes wrong

## 🧪 Test the Build

```powershell
powershell -ExecutionPolicy Bypass -File test-plugin.ps1
```

Runs automated tests to verify:
- Plugin DLL was created
- Source files are complete
- Documentation is updated
- Dead code is removed

## 📦 Create Release Package

```powershell
.\create-release.bat
```

Creates a distributable ZIP file containing:
- Plugin DLL
- Installation scripts
- Documentation
- Quick start guide

## 🔧 Troubleshooting

If build fails, see `BUILD_TROUBLESHOOTING.md` for:
- Common error solutions
- CMake configuration help
- Qt6 setup instructions
- Clean build procedures

## 📋 Complete Workflow

```powershell
# 1. Build
.\build-quick.bat

# 2. Test
powershell -ExecutionPolicy Bypass -File test-plugin.ps1

# 3. Install to OBS
.\install.bat

# 4. Test in OBS
# (manually add "QR Donations" source and test)

# 5. Create release
.\create-release.bat
```

## 🎯 What's New

### Automation Scripts
1. **build-quick.bat** - One-click build with Qt auto-detection
2. **test-plugin.ps1** - Automated test suite
3. **create-release.bat** - Release package creator

### Documentation
1. **BUILD_TROUBLESHOOTING.md** - Fix common build issues
2. **BUILD_QUICK_START.md** - This file

## ✅ Success Indicators

After running `build-quick.bat`, you should see:
```
[OK] CMake found
[OK] Qt6 found at: C:\Qt\6.x\msvc2019_64
[OK] Configuration successful
[OK] Build complete!
```

Plugin location: `build\Release\obs-qr-donations.dll`

## 🐛 Common Issues

| Issue | Quick Fix |
|-------|-----------|
| CMake not found | `choco install cmake` or download from cmake.org |
| Qt6 not found | Install Qt6 from qt.io, then edit build-quick.bat with path |
| OBS SDK missing | Plugin uses stub mode by default, should work without OBS SDK |
| Build fails | See `BUILD_TROUBLESHOOTING.md` |

## 📞 Need Help?

1. Check `BUILD_TROUBLESHOOTING.md`
2. Run test script to identify issues
3. Review OBS logs after installation
4. Check error messages in build output

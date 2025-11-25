# OBS QR Donations Plugin - Testing Scripts

This directory contains automated testing and build scripts for the OBS QR Donations plugin.

## Quick Start

### Windows
```powershell
# Build and test
.\scripts\build-and-test.bat

# Or use PowerShell script with more options
.\scripts\quick-build.ps1

# Skip tests
.\scripts\quick-build.ps1 -SkipTests

# Build with full Breez SDK (not stub)
.\scripts\quick-build.ps1 -FullSDK
```

### Linux/macOS
```bash
# Make scripts executable
chmod +x scripts/*.sh

# Build and test
./scripts/build-and-test.sh

# Skip tests
./scripts/build-and-test.sh --skip-tests

# Build with full Breez SDK
./scripts/build-and-test.sh --full-sdk
```

## Available Scripts

### Build Scripts

- **`quick-build.ps1`** (Windows PowerShell)
  - Comprehensive build script with prerequisite checking
  - Optional test execution
  - Supports both stub and full SDK modes
  - Usage: `.\scripts\quick-build.ps1 [-SkipTests] [-FullSDK]`

- **`build-and-test.bat`** (Windows Batch)
  - Simple batch script for quick builds
  - Automatically runs tests if Python is available
  - Usage: `.\scripts\build-and-test.bat`

- **`build-and-test.sh`** (Linux/macOS)
  - Shell script with prerequisite checking
  - Colored output for easy reading
  - Usage: `./scripts/build-and-test.sh [--skip-tests] [--full-sdk]`

### Test Scripts

- **`run_integration_test.py`**
  - Comprehensive integration test suite
  - Tests plugin binary, QR generation, configuration
  - Usage: `python scripts/run_integration_test.py [--build-dir BUILD_DIR]`

- **`test_plugin.py`**
  - Interactive GUI for manual testing
  - Generate and test QR codes
  - Requires PyQt6 and qrcode libraries
  - Usage: `python scripts/test_plugin.py`

- **`test_installation.py`**
  - Verifies plugin installation
  - Checks for required files and OBS compatibility
  - Usage: `python scripts/test_installation.py`

### Other Scripts

- **`setup_spark_sdk.bat`** / **`setup_spark_sdk.sh`**
  - Downloads and builds the Breez Spark SDK
  - Required only for full Lightning support
  - Usage: See BUILD_GUIDE.md

## Test Coverage

The integration test suite covers:

1. **File Structure** - Verifies all required files exist
2. **Plugin Binary** - Checks if the plugin compiled successfully
3. **Stub Configuration** - Validates CMake configuration
4. **Dependencies** - Ensures Python testing libraries are available
5. **QR Generation** - Tests QR code creation functionality

## Requirements

### For Building
- CMake 3.16+
- Qt6
- C++ compiler (MSVC 2019+, GCC 9+, or Clang 10+)
- OBS Studio development files

### For Testing
- Python 3.7+
- Python packages:
  ```bash
  pip install qrcode pillow PyQt6
  ```

## Continuous Integration

These scripts are designed to work in CI/CD environments:

```yaml
# Example GitHub Actions workflow
- name: Build and Test
  run: |
    ./scripts/build-and-test.sh --skip-tests
    python scripts/run_integration_test.py
```

## Troubleshooting

If build fails:
1. Check that all prerequisites are installed (see BUILD_GUIDE.md)
2. Verify OBS Studio development files are available
3. Ensure Qt6 is in your PATH
4. Try building in stub mode first: `-DBREEZ_USE_STUB=ON`

If tests fail:
1. Install Python dependencies: `pip install -r requirements-dev.txt`
2. Check that the plugin binary exists in `build/Release/` or `build/`
3. Review test output for specific failures

## See Also

- [BUILD_GUIDE.md](../BUILD_GUIDE.md) - Complete build instructions
- [README.md](../README.md) - Project overview
- [USER_GUIDE.md](../USER_GUIDE.md) - End-user documentation

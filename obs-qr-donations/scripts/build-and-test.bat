@echo off
REM ==================================================================
REM OBS QR Donations Plugin - Windows Build and Test Script
REM ==================================================================

echo.
echo ========================================
echo OBS QR Donations - Build ^& Test
echo ========================================
echo.

REM Step 1: Check prerequisites
echo [1/5] Checking prerequisites...
echo.

where cmake >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: CMake not found. Please install CMake.
    goto :error
)
echo   [✓] CMake found

where git >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: Git not found. Please install Git.
    goto :error
)
echo   [✓] Git found

echo.
echo [2/5] Configuring CMake build...
echo.

cmake -B build -S . -DBREEZ_USE_STUB=ON -DBREEZ_STUB_SIMULATE=ON
if %ERRORLEVEL% neq 0 (
    echo.
    echo ERROR: CMake configuration failed.
    echo.
    echo Troubleshooting tips:
    echo   - Make sure OBS Studio is installed
    echo   - Install Qt6 from https://www.qt.io/download
    echo   - Check BUILD_GUIDE.md for detailed instructions
    goto :error
)
echo   [✓] Configuration successful

echo.
echo [3/5] Building plugin...
echo.

cmake --build build --config Release
if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed.
    goto :error
)
echo   [✓] Build successful

echo.
echo [4/5] Verifying build...
echo.

if exist "build\Release\obs-qr-donations.dll" (
    echo   [✓] Plugin binary found
    dir /b build\Release\obs-qr-donations.dll
) else (
    echo ERROR: Plugin binary not found
    goto :error
)

echo.
echo [5/5] Running tests...
echo.

where python >nul 2>&1
if %ERRORLEVEL% equ 0 (
    python scripts\run_integration_test.py
) else (
    echo   [!] Python not found, skipping tests
)

echo.
echo ========================================
echo BUILD COMPLETE!
echo ========================================
echo.
echo Plugin location: build\Release\obs-qr-donations.dll
echo.
echo To install:
echo   1. Run install.bat
echo   2. Or manually copy to: C:\Program Files\obs-studio\obs-plugins\64bit\
echo.

goto :end

:error
echo.
echo ========================================
echo BUILD FAILED
echo ========================================
echo.
echo Please check the error messages above and refer to BUILD_GUIDE.md
echo.
exit /b 1

:end
echo Press any key to exit...
pause >nul

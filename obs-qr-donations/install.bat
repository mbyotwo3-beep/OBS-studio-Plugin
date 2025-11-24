@echo off
REM OBS QR Donations Plugin - Easy Installer
REM This script helps install the plugin to OBS Studio

echo ========================================
echo  OBS QR Donations Plugin - Installer
echo ========================================
echo.

REM Find OBS Studio installation
set "OBS_PATH="
if exist "C:\Program Files\obs-studio\" set "OBS_PATH=C:\Program Files\obs-studio"
if exist "C:\Program Files (x86)\obs-studio\" set "OBS_PATH=C:\Program Files (x86)\obs-studio"
if exist "%APPDATA%\obs-studio\" set "OBS_PATH=%APPDATA%\obs-studio"

if "%OBS_PATH%"=="" (
    echo [ERROR] Could not find OBS Studio installation
    echo Please install OBS Studio from: https://obsproject.com/
    pause
    exit /b 1
)

echo [OK] Found OBS Studio at: %OBS_PATH%
echo.

REM Create plugin directories
set "PLUGIN_DIR=%OBS_PATH%\obs-plugins\64bit"
set "DATA_DIR=%OBS_PATH%\data\obs-plugins\obs-qr-donations"

echo Creating plugin directories...
if not exist "%PLUGIN_DIR%" mkdir "%PLUGIN_DIR%"
if not exist "%DATA_DIR%" mkdir "%DATA_DIR%"

REM Check if plugin DLL exists in build directory
if not exist "build\Release\obs-qr-donations.dll" (
    echo [WARNING] Plugin not built yet!
    echo.
    echo Please build the plugin first:
    echo   1. Install dependencies (see README.md)
    echo   2. Run: cmake -B build -S . -DBREEZ_USE_STUB=ON
    echo   3. Run: cmake --build build --config Release
    echo   4. Run this installer again
    echo.
    pause
    exit /b 1
)

REM Copy plugin files
echo Installing plugin files...
copy /Y "build\Release\obs-qr-donations.dll" "%PLUGIN_DIR%\" >nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to copy plugin DLL. Try running as Administrator.
    pause
    exit /b 1
)

REM Copy data files if they exist
if exist "data\" (
    xcopy /E /I /Y "data\*" "%DATA_DIR%\" >nul
)

echo.
echo ========================================
echo  Installation Complete!
echo ========================================
echo.
echo Plugin installed to: %PLUGIN_DIR%\obs-qr-donations.dll
echo Data files at: %DATA_DIR%
echo.
echo To use the plugin:
echo   1. Open OBS Studio
echo   2. Click + in Sources panel
echo   3. Select "QR Donations"
echo   4. Configure your cryptocurrency addresses
echo.
echo For Lightning Network support:
echo   - Get a free Breez API key from https://breez.technology/
echo   - Enter it in the plugin settings
echo.
pause

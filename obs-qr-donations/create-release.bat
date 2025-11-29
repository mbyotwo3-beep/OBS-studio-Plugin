@echo off
REM Create Release Package for OBS QR Donations Plugin
REM This script packages the plugin for distribution

echo ============================================
echo  OBS QR Donations - Package Creator
echo ============================================
echo.

REM Check if plugin is built
if not exist "build\Release\obs-qr-donations.dll" (
    echo [ERROR] Plugin not built yet!
    echo Please run build-quick.bat first
    pause
    exit /b 1
)

echo [OK] Plugin DLL found
echo.

REM Get version from CMakeLists.txt
set VERSION=1.0.0

REM Create package directory
set PACKAGE_DIR=obs-qr-donations-v%VERSION%-windows
if exist %PACKAGE_DIR% (
    echo Cleaning old package...
    rmdir /s /q %PACKAGE_DIR%
)

echo Creating package directory: %PACKAGE_DIR%
mkdir %PACKAGE_DIR%
mkdir %PACKAGE_DIR%\plugin
mkdir %PACKAGE_DIR%\docs

echo.
echo Copying files...
echo.

REM Copy plugin DLL
echo - Plugin DLL
copy build\Release\obs-qr-donations.dll %PACKAGE_DIR%\plugin\

REM Copy installation scripts
echo - Installation scripts
copy install.bat %PACKAGE_DIR%\
copy install.sh %PACKAGE_DIR%\

REM Copy documentation
echo - Documentation
copy README.md %PACKAGE_DIR%\docs\
copy USER_GUIDE.md %PACKAGE_DIR%\docs\
copy QUICK_START.md %PACKAGE_DIR%\docs\
copy LICENSE %PACKAGE_DIR%\docs\

REM Create README for the package
echo Creating package README...
(
echo # OBS QR Donations Plugin v%VERSION%
echo.
echo ## Quick Install
echo.
echo ### Windows
echo 1. Run `install.bat`
echo 2. Restart OBS
echo 3. Add "QR Donations" source
echo.
echo ### Linux/Mac
echo ```bash
echo chmod +x install.sh
echo ./install.sh
echo ```
echo.
echo ## Documentation
echo - **Quick Start Guide**: docs/QUICK_START.md
echo - **Full User Guide**: docs/USER_GUIDE.md
echo - **README**: docs/README.md
echo.
echo ## Features
echo - ⚡ Lightning Network ^(Breez SDK^)
echo - 🎆 Visual flash effects on donation
echo - 🔔 Optional audio notifications
echo - 📱 Fully responsive display
echo.
echo ## Support
echo See USER_GUIDE.md for troubleshooting and setup help.
) > %PACKAGE_DIR%\README.txt

echo.
echo Creating ZIP archive...
powershell -Command "Compress-Archive -Path '%PACKAGE_DIR%' -DestinationPath '%PACKAGE_DIR%.zip' -Force"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ============================================
    echo  Package Complete!
    echo ============================================
    echo.
    echo Package: %PACKAGE_DIR%.zip
    echo Size: 
    dir %PACKAGE_DIR%.zip | find ".zip"
    echo.
    echo Ready for distribution!
    echo.
) else (
    echo [ERROR] Failed to create ZIP archive
)

pause

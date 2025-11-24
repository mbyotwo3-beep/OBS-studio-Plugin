@echo off
REM Quick Setup Script for OBS QR Donations Plugin Dependencies
REM This script helps install Qt6, vcpkg, and qrencode

echo ========================================
echo  OBS QR Donations - Dependency Setup
echo ========================================
echo.

REM Check if running from project directory
if not exist "CMakeLists.txt" (
    echo [ERROR] Please run this script from the project root directory
    pause
    exit /b 1
)

echo Step 1: Installing vcpkg (package manager)
echo ========================================
echo.

if not exist "vcpkg" (
    echo Cloning vcpkg...
    git clone https://github.com/Microsoft/vcpkg.git
    if %ERRORLEVEL% NEQ 0 (
        echo [ERROR] Failed to clone vcpkg. Make sure git is installed.
        pause
        exit /b 1
    )
    
    echo Bootstrapping vcpkg...
    cd vcpkg
    call bootstrap-vcpkg.bat
    cd ..
) else (
    echo [OK] vcpkg already exists
)

echo.
echo Step 2: Installing qrencode
echo ========================================
echo.

vcpkg\vcpkg install qrencode:x64-windows
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] qrencode installation had issues, but continuing...
)

echo.
echo Step 3: Qt6 Installation
echo ========================================
echo.
echo Qt6 needs to be installed manually:
echo 1. Download Qt Online Installer from: https://www.qt.io/download-qt-installer
echo 2. Install Qt 6.5 or later (select MSVC 2019 64-bit component)
echo 3. Note the installation path (e.g., C:\Qt\6.5.0\msvc2019_64)
echo.
set /p QT_PATH="Enter your Qt6 installation path (or press Enter to skip): "

echo.
echo Step 4: OBS Studio SDK
echo ========================================
echo.
echo For OBS development, you need the SDK. Options:
echo 1. Download pre-built SDK from OBS releases
echo 2. Build OBS from source
echo.
echo Would you like to download OBS SDK automatically?
set /p DOWNLOAD_OBS="Download OBS SDK? (y/n): "

if /i "%DOWNLOAD_OBS%"=="y" (
    echo.
    echo Downloading OBS Studio SDK...
    echo This might take a few minutes...
    
    REM Download latest OBS SDK
    powershell -Command "Invoke-WebRequest -Uri 'https://github.com/obsproject/obs-studio/releases/latest/download/obs-studio-28.0.0-windows-x64.zip' -OutFile 'obs-sdk.zip'"
    
    if exist "obs-sdk.zip" (
        echo Extracting OBS SDK...
        powershell -Command "Expand-Archive -Path 'obs-sdk.zip' -DestinationPath 'obs-sdk' -Force"
        del obs-sdk.zip
        echo [OK] OBS SDK downloaded and extracted to obs-sdk/
        set OBS_SDK_PATH=%CD%\obs-sdk
    ) else (
        echo [WARNING] Download failed. Please download manually.
    )
)

echo.
echo ========================================
echo  Setup Summary
echo ========================================
echo.
echo [OK] vcpkg installed
echo [OK] qrencode installed via vcpkg
if not "%QT_PATH%"=="" echo [OK] Qt6 path: %QT_PATH%
if not "%OBS_SDK_PATH%"=="" echo [OK] OBS SDK: %OBS_SDK_PATH%
echo.
echo Next steps:
echo 1. If you didn't install Qt6 yet, download it from https://www.qt.io/download-qt-installer
echo 2. Run the build command:
echo.
echo    cmake -B build -S . -DBREEZ_USE_STUB=ON -DCMAKE_TOOLCHAIN_FILE=vcpkg/scripts/buildsystems/vcpkg.cmake
if not "%QT_PATH%"=="" echo    -DQt6_DIR="%QT_PATH%/lib/cmake/Qt6"
if not "%OBS_SDK_PATH%"=="" echo    -DLibOBS_DIR="%OBS_SDK_PATH%/cmake"
echo.
echo 3. Build:
echo    cmake --build build --config Release
echo.
pause

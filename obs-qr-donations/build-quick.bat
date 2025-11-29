@echo off
REM Quick Build Script for OBS QR Donations Plugin
REM This script attempts to configure and build the plugin

echo ============================================
echo  OBS QR Donations - Build Script
echo ============================================
echo.

REM Check for CMake
where cmake >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] CMake not found in PATH
    echo Please install CMake or add it to your PATH
    pause
    exit /b 1
)

echo [OK] CMake found
echo.

REM Determine Qt6 path
set QT_SEARCH_PATHS=C:\Qt C:\Qt6 "C:\Program Files\Qt"
set QT6_PATH=

for %%P in (%QT_SEARCH_PATHS%) do (
    if exist "%%P" (
        for /d %%V in ("%%P\6.*") do (
            if exist "%%V\msvc2019_64\lib\cmake\Qt6" (
                set QT6_PATH=%%V\msvc2019_64
                goto :qt_found
            )
            if exist "%%V\msvc2022_64\lib\cmake\Qt6" (
                set QT6_PATH=%%V\msvc2022_64
                goto :qt_found
            )
        )
    )
)

:qt_found
if "%QT6_PATH%"=="" (
    echo [WARNING] Qt6 not found automatically
    echo.
    set /p QT6_PATH="Enter Qt6 path (e.g., C:\Qt\6.5.3\msvc2019_64): "
) else (
    echo [OK] Qt6 found at: %QT6_PATH%
)

echo.
echo ============================================
echo  Configuring Build
echo ============================================
echo.

REM Check if build directory exists
if exist build (
    echo Cleaning old build directory...
    rmdir /s /q build
)

REM Configure with CMake
cmake -B build -S . ^
    -DBREEZ_USE_STUB=ON ^
    -DQt6_DIR="%QT6_PATH%/lib/cmake/Qt6" 

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] CMake configuration failed
    echo.
    echo Possible issues:
    echo - Qt6 path incorrect
    echo - Missing dependencies
    echo - OBS SDK not found
    echo.
    pause
    exit /b 1
)

echo.
echo [OK] Configuration successful
echo.
echo ============================================
echo  Building Plugin
echo ============================================
echo.

cmake --build build --config Release

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] Build failed
    echo Check the error messages above
    pause
    exit /b 1
)

echo.
echo ============================================
echo  Build Complete!
echo ============================================
echo.
echo Plugin built successfully!
echo Location: build\Release\obs-qr-donations.dll
echo.
echo Next steps:
echo 1. Run install.bat to install to OBS
echo 2. Restart OBS
echo 3. Add "QR Donations" source to test
echo.
pause

@echo off
REM Manual Build Script - For when Qt6 can't be auto-detected

echo ============================================
echo  OBS QR Donations - Manual Build
echo ============================================
echo.

echo This script will help you build the plugin manually.
echo.

REM Check CMake
where cmake >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] CMake not found
    echo Install from: https://cmake.org/download/
    pause
    exit /b 1
)

echo [OK] CMake found
echo.

echo ============================================
echo  Qt6 Path Required
echo ============================================
echo.
echo Common Qt6 locations:
echo - C:\Qt\6.5.3\msvc2019_64
echo - C:\Qt\6.6.0\msvc2022_64
echo - C:\Qt6\6.5.3\msvc2019_64
echo - C:\Program Files\Qt\6.5.3\msvc2019_64
echo.
echo The path should contain: lib\cmake\Qt6
echo.

set /p QT6_PATH="Enter Qt6 path: "

if "%QT6_PATH%"=="" (
    echo [ERROR] Qt6 path required
    pause
    exit /b 1
)

if not exist "%QT6_PATH%\lib\cmake\Qt6" (
    echo [WARNING] Qt6 path may be incorrect
    echo Expected to find: %QT6_PATH%\lib\cmake\Qt6
    echo.
    set /p CONTINUE="Continue anyway? (y/n): "
    if /i not "%CONTINUE%"=="y" exit /b 1
)

echo.
echo Using Qt6: %QT6_PATH%
echo.

REM Clean build directory
if exist build (
    echo Cleaning build directory...
    rmdir /s /q build
)

echo ============================================
echo  Configuring with CMake
echo ============================================
echo.

cmake -B build -S . ^
    -DBREEZ_USE_STUB=ON ^
    -DQt6_DIR="%QT6_PATH%/lib/cmake/Qt6"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] Configuration failed
    echo.
    echo Troubleshooting:
    echo 1. Verify Qt6 path is correct
    echo 2. Check that Qt Multimedia is installed
    echo 3. See BUILD_TROUBLESHOOTING.md
    echo.
    pause
    exit /b 1
)

echo.
echo [OK] Configuration successful
echo.

echo ============================================
echo  Building
echo ============================================
echo.

cmake --build build --config Release

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] Build failed
    echo See error messages above
    echo.
    pause
    exit /b 1
)

echo.
echo ============================================
echo  BUILD SUCCESS!
echo ============================================
echo.
echo Plugin: build\Release\obs-qr-donations.dll
echo.
echo Next steps:
echo 1. Run: install.bat
echo 2. Restart OBS
echo 3. Add "QR Donations" source
echo.
pause

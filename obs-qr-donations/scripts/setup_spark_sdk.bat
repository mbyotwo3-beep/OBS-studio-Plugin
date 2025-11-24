@echo off
REM Setup script for Breez SDK Spark (Windows)

echo === Breez SDK Spark Setup Script ===
echo.

REM Check for Rust/Cargo
where cargo >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Error: Rust/Cargo not found!
    echo Please install Rust from: https://rustup.rs/
    exit /b 1
)

echo [OK] Rust/Cargo found
echo.

REM Navigate to SDK directory
set SCRIPT_DIR=%~dp0
set SDK_DIR=%SCRIPT_DIR%..\third_party\breez_sdk

if not exist "%SDK_DIR%" (
    echo Error: Breez SDK not found at %SDK_DIR%
    echo Please run: git clone https://github.com/breez/spark-sdk.git third_party/breez_sdk
    exit /b 1
)

cd /d "%SDK_DIR%"
echo Building Breez SDK Spark...
echo This may take several minutes...
echo.

REM Build the SDK with release optimizations
cargo build --release --package breez-sdk-core

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Error: Build failed!
    exit /b 1
)

echo.
echo === Build Complete ===
echo.
echo Library location: %SDK_DIR%\target\release\
echo.
echo Next steps:
echo 1. Configure CMake with: cmake -B build -S . -DCMAKE_TOOLCHAIN_FILE=..\vcpkg\scripts\buildsystems\vcpkg.cmake
echo 2. Build the OBS plugin: cmake --build build --config Release
echo.

pause

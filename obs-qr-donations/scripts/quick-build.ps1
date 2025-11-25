# OBS QR Donations - Build & Test Automation

## Quick Build Script for Windows

This script automates the complete build, test, and deployment process.

```powershell
# quick-build.ps1
param(
    [switch]$SkipTests,
    [switch]$FullSDK,
    [string]$ObsPath = "C:\Program Files\obs-studio"
)

Write-Host "🚀 OBS QR Donations - Quick Build Script" -ForegroundColor Cyan
Write-Host "========================================`n"

# Step 1: Check prerequisites
Write-Host "📋 Step 1: Checking prerequisites..." -ForegroundColor Yellow

$prerequisites = @{
    "CMake" = "cmake --version"
    "Git" = "git --version"
    "Qt6" = "qmake --version"
}

$missingPrereqs = @()
foreach ($tool in $prerequisites.Keys) {
    try {
        $null = Invoke-Expression $prerequisites[$tool] 2>&1
        Write-Host "  ✅ $tool found" -ForegroundColor Green
    } catch {
        Write-Host "  ❌ $tool not found" -ForegroundColor Red
        $missingPrereqs += $tool
    }
}

if ($missingPrereqs.Count -gt 0) {
    Write-Host "`n❌ Missing prerequisites: $($missingPrereqs -join ', ')" -ForegroundColor Red
    Write-Host "Please install the missing tools and try again." -ForegroundColor Yellow
    Write-Host "See BUILD_GUIDE.md for installation instructions.`n"
    exit 1
}

# Step 2: Configure CMake
Write-Host "`n📦 Step 2: Configuring CMake build..." -ForegroundColor Yellow

$cmakeArgs = @(
    "-B", "build",
    "-S", ".",
    "-DBREEZ_USE_STUB=ON",
    "-DBREEZ_STUB_SIMULATE=ON"
)

if ($FullSDK) {
    Write-Host "  Building with full Breez SDK..." -ForegroundColor Cyan
    $cmakeArgs = $cmakeArgs | Where-Object { $_ -notmatch "BREEZ_USE_STUB" }
}

try {
    & cmake @cmakeArgs
    if ($LASTEXITCODE -ne 0) {
        throw "CMake configuration failed"
    }
    Write-Host "  ✅ CMake configuration successful" -ForegroundColor Green
} catch {
    Write-Host "  ❌ CMake configuration failed" -ForegroundColor Red
    Write-Host "  Error: $_" -ForegroundColor Red
    Write-Host "`n💡 Troubleshooting tips:" -ForegroundColor Yellow
    Write-Host "  - Make sure OBS Studio is installed"
    Write-Host "  - Check that Qt6 is in your PATH"
    Write-Host "  - Try specifying paths manually:"
    Write-Host "    cmake -DLibOBS_DIR='path/to/obs' -DQt6_DIR='path/to/qt/cmake/Qt6'"
    exit 1
}

# Step 3: Build the plugin
Write-Host "`n🔨 Step 3: Building plugin..." -ForegroundColor Yellow

try {
    & cmake --build build --config Release
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed"
    }
    Write-Host "  ✅ Build successful" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Build failed" -ForegroundColor Red
    Write-Host "  Error: $_" -ForegroundColor Red
    exit 1
}

# Step 4: Run tests
if (-not $SkipTests) {
    Write-Host "`n🧪 Step 4: Running tests..." -ForegroundColor Yellow
    
    # Check if plugin binary exists
    $pluginPath = "build\Release\obs-qr-donations.dll"
    if (Test-Path $pluginPath) {
        Write-Host "  ✅ Plugin binary found: $pluginPath" -ForegroundColor Green
        
        # Get file info
        $fileInfo = Get-Item $pluginPath
        Write-Host "  📊 Size: $([math]::Round($fileInfo.Length / 1KB, 2)) KB" -ForegroundColor Cyan
        Write-Host "  📅 Modified: $($fileInfo.LastWriteTime)" -ForegroundColor Cyan
    } else {
        Write-Host "  ❌ Plugin binary not found" -ForegroundColor Red
        exit 1
    }
    
    # Run Python tests if available
    if (Get-Command python -ErrorAction SilentlyContinue) {
        Write-Host "`n  Running plugin tests..." -ForegroundColor Cyan
        try {
            python scripts\test_plugin.py
            Write-Host "  ✅ Tests passed" -ForegroundColor Green
        } catch {
            Write-Host "  ⚠️  Tests failed or skipped" -ForegroundColor Yellow
        }
    }
} else {
    Write-Host "`n⏭️  Step 4: Skipping tests (--SkipTests specified)" -ForegroundColor Yellow
}

# Step 5: Installation
Write-Host "`n📦 Step 5: Installation options..." -ForegroundColor Yellow
Write-Host "  Plugin built successfully at: build\Release\obs-qr-donations.dll" -ForegroundColor Cyan
Write-Host "`n  To install:" -ForegroundColor Yellow
Write-Host "    Option 1: Run .\install.bat" -ForegroundColor Cyan
Write-Host "    Option 2: Manually copy to: $ObsPath\obs-plugins\64bit\" -ForegroundColor Cyan

Write-Host "`n🎉 Build process complete!" -ForegroundColor Green
Write-Host "========================================`n"

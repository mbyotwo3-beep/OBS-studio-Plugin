# OBS Plugin Build Using vcpkg (Simplified Approach)
# This script automates the entire build process including dependencies

Write-Host "`n🚀 OBS QR Donations - Automated vcpkg Build" -ForegroundColor Cyan
Write-Host "=" * 60 -ForegroundColor Cyan

# Step 1: Check if vcpkg is installed
$vcpkgPath = "C:\vcpkg"
$vcpkgInstalled = Test-Path $vcpkgPath

if (-not $vcpkgInstalled) {
    Write-Host "`n[1/5] Installing vcpkg package manager..." -ForegroundColor Yellow
    Write-Host "  This will download and set up vcpkg (one-time setup)" -ForegroundColor Cyan
    
    try {
        git clone https://github.com/Microsoft/vcpkg.git $vcpkgPath
        Set-Location $vcpkgPath
        .\bootstrap-vcpkg.bat
        
        if ($LASTEXITCODE -ne 0) {
            throw "vcpkg bootstrap failed"
        }
        
        Write-Host "  ✅ vcpkg installed successfully" -ForegroundColor Green
    }
    catch {
        Write-Host "  ❌ Failed to install vcpkg: $_" -ForegroundColor Red
        exit 1
    }
}
else {
    Write-Host "`n[1/5] vcpkg already installed" -ForegroundColor Green
}

# Step 2: Install dependencies via vcpkg
Write-Host "`n[2/5] Installing dependencies (Qt6, qrencode)..." -ForegroundColor Yellow
Write-Host "  ⏳ This may take 20-30 minutes on first run..." -ForegroundColor Cyan

Set-Location $vcpkgPath

$packages = @("qt6-base:x64-windows", "qrencode:x64-windows")
foreach ($package in $packages) {
    Write-Host "  Installing $package..." -ForegroundColor Cyan
    .\vcpkg install $package
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ❌ Failed to install $package" -ForegroundColor Red
        exit 1
    }
}

Write-Host "  ✅ Dependencies installed" -ForegroundColor Green

# Step 3: Navigate back to plugin directory
$pluginDir = "C:\Users\Administrator\Desktop\OBS studio Plugin\obs-qr-donations"
Set-Location $pluginDir

# Step 4: Configure CMake with vcpkg toolchain
Write-Host "`n[3/5] Configuring CMake build..." -ForegroundColor Yellow

$toolchainFile = "$vcpkgPath\scripts\buildsystems\vcpkg.cmake"

try {
    cmake -B build -S . `
        -DBREEZ_USE_STUB=ON `
        -DBREEZ_STUB_SIMULATE=ON `
        -DCMAKE_TOOLCHAIN_FILE="$toolchainFile" `
        -G "Visual Studio 17 2022"
    
    if ($LASTEXITCODE -ne 0) {
        throw "CMake configuration failed"
    }
    
    Write-Host "  ✅ CMake configuration successful" -ForegroundColor Green
}
catch {
    Write-Host "  ❌ CMake configuration failed" -ForegroundColor Red
    Write-Host "  Note: OBS Studio SDK still required for plugin compilation" -ForegroundColor Yellow
    Write-Host "  You can use the demo and test scripts instead!" -ForegroundColor Cyan
    exit 1
}

# Step 5: Build the plugin
Write-Host "`n[4/5] Building plugin..." -ForegroundColor Yellow

try {
    cmake --build build --config Release
    
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed"
    }
    
    Write-Host "  ✅ Build successful!" -ForegroundColor Green
}
catch {
    Write-Host "  ❌ Build failed" -ForegroundColor Red
    exit 1
}

# Step 6: Verify build output
Write-Host "`n[5/5] Verifying build..." -ForegroundColor Yellow

$pluginDll = "build\Release\obs-qr-donations.dll"
if (Test-Path $pluginDll) {
    $fileInfo = Get-Item $pluginDll
    Write-Host "  ✅ Plugin built successfully!" -ForegroundColor Green
    Write-Host "  📁 Location: $pluginDll" -ForegroundColor Cyan
    Write-Host "  📊 Size: $([math]::Round($fileInfo.Length / 1KB, 2)) KB" -ForegroundColor Cyan
    
    Write-Host "`n🎉 BUILD COMPLETE!" -ForegroundColor Green
    Write-Host "=" * 60 -ForegroundColor Green
    Write-Host "`nNext steps:" -ForegroundColor Yellow
    Write-Host "  1. Run .\install.bat to install to OBS" -ForegroundColor Cyan
    Write-Host "  2. Start OBS Studio" -ForegroundColor Cyan
    Write-Host "  3. Add source: Sources → + → QR Donations" -ForegroundColor Cyan
}
else {
    Write-Host "  ❌ Plugin binary not found" -ForegroundColor Red
    exit 1
}

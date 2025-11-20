<#
.SYNOPSIS
    Tests the local installation of the OBS QR Donations plugin.
.DESCRIPTION
    This script verifies that the plugin is correctly installed in the OBS Studio plugins directory
    and that all required files are present.
#>

# Configuration
$OBS_PLUGIN_DIR = "$env:APPDATA\obs-studio\plugins"
$PLUGIN_NAME = "qr-donations"
$PLUGIN_DIR = Join-Path $OBS_PLUGIN_DIR $PLUGIN_NAME

# Check if running as administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Host "This script requires administrator privileges to check system directories." -ForegroundColor Yellow
    Write-Host "Please run PowerShell as Administrator and try again." -ForegroundColor Yellow
    exit 1
}

Write-Host "🔍 Checking OBS QR Donations plugin installation..." -ForegroundColor Cyan

# Check if OBS plugin directory exists
if (-not (Test-Path $OBS_PLUGIN_DIR)) {
    Write-Host "❌ OBS Studio plugins directory not found at: $OBS_PLUGIN_DIR" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Found OBS Studio plugins directory: $OBS_PLUGIN_DIR" -ForegroundColor Green

# Check if plugin directory exists
if (-not (Test-Path $PLUGIN_DIR)) {
    Write-Host "❌ Plugin directory not found at: $PLUGIN_DIR" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Found plugin directory: $PLUGIN_DIR" -ForegroundColor Green

# Check for required files
$requiredFiles = @(
    "bin\obs-qr-donations.dll",
    "data\settings.json",
    "README.txt"
)

$allFilesExist = $true
foreach ($file in $requiredFiles) {
    $filePath = Join-Path $PLUGIN_DIR $file
    if (-not (Test-Path $filePath)) {
        Write-Host "❌ Missing required file: $file" -ForegroundColor Red
        $allFilesExist = $false
    } else {
        Write-Host "✅ Found: $file" -ForegroundColor Green
    }
}

if (-not $allFilesExist) {
    Write-Host "❌ Some required files are missing. Please reinstall the plugin." -ForegroundColor Red
    exit 1
}

# Check OBS version
$obsExePath = "${env:ProgramFiles}\obs-studio\bin\64bit\obs64.exe"
if (Test-Path $obsExePath) {
    $versionInfo = (Get-Item $obsExePath).VersionInfo
    $version = "$($versionInfo.FileMajorPart).$($versionInfo.FileMinorPart).$($versionInfo.FileBuildPart)"
    
    Write-Host "✅ Found OBS Studio version: $version" -ForegroundColor Green
    
    # Check if version is 28.0 or later
    $majorVersion = [int]$versionInfo.FileMajorPart
    if ($majorVersion -lt 28) {
        Write-Host "⚠️  Warning: OBS Studio version $version may not be compatible. Version 28.0 or later is recommended." -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  Could not determine OBS Studio version. Make sure it's installed." -ForegroundColor Yellow
}

# Final check
if ($allFilesExist) {
    Write-Host "
🎉 OBS QR Donations plugin is properly installed!" -ForegroundColor Green
    Write-Host "Restart OBS Studio to see the plugin in the sources list." -ForegroundColor Cyan
} else {
    Write-Host "
❌ Installation check failed. Please check the errors above." -ForegroundColor Red
    exit 1
}

# Open OBS plugin directory in Explorer
$openDir = Read-Host "Do you want to open the plugin directory in Explorer? (y/n)"
if ($openDir -eq 'y') {
    explorer $PLUGIN_DIR
}

# Check if OBS is running and offer to restart it
$obsProcess = Get-Process obs64 -ErrorAction SilentlyContinue
if ($obsProcess) {
    $restartObs = Read-Host "OBS Studio is running. Do you want to restart it to apply changes? (y/n)"
    if ($restartObs -eq 'y') {
        Stop-Process -Name obs64 -Force
        Start-Process $obsExePath
    }
}

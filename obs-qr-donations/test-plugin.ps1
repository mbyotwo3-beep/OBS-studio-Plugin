# OBS QR Donations Plugin - Test Script
# Automated testing procedures

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " OBS QR Donations - Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$testsPassed = 0
$testsFailed = 0

function Test-Feature {
    param(
        [string]$Name,
        [scriptblock]$Test
    )
    
    Write-Host "Testing: $Name..." -NoNewline
    try {
        $result = & $Test
        if ($result) {
            Write-Host " [PASS]" -ForegroundColor Green
            $script:testsPassed++
        }
        else {
            Write-Host " [FAIL]" -ForegroundColor Red
            $script:testsFailed++
        }
    }
    catch {
        Write-Host " [ERROR]" -ForegroundColor Red
        Write-Host "  Error: $_" -ForegroundColor Red
        $script:testsFailed++
    }
}

Write-Host "1. Build Verification Tests" -ForegroundColor Yellow
Write-Host "----------------------------" -ForegroundColor Yellow

Test-Feature "Plugin DLL exists" {
    Test-Path "build\Release\obs-qr-donations.dll"
}

Test-Feature "Source files present" {
    (Test-Path "src\qr-donations.cpp") -and 
    (Test-Path "src\qr-widget.cpp") -and
    (Test-Path "src\breez-service.cpp")
}

Test-Feature "Headers present" {
    (Test-Path "src\qr-donations.hpp") -and
    (Test-Path "src\qr-widget.hpp") -and
    (Test-Path "src\breez-service.hpp")
}

Test-Feature "CMakeLists.txt valid" {
    $content = Get-Content "CMakeLists.txt" -Raw
    $content -match "Qt6::Multimedia" -and
    $content -match "obs-qr-donations"
}

Write-Host ""
Write-Host "2. Code Quality Tests" -ForegroundColor Yellow
Write-Host "---------------------" -ForegroundColor Yellow

Test-Feature "No donation-effect dead code" {
    -not (Test-Path "src\donation-effect.cpp") -and
    -not (Test-Path "src\donation-effect.hpp")
}

Test-Feature "QSoundEffect included" {
    $content = Get-Content "src\qr-donations.hpp" -Raw
    $content -match "QSoundEffect"
}

Test-Feature "Flash overlay implemented" {
    $content = Get-Content "src\qr-widget.cpp" -Raw
    $content -match "flashOverlay" -and
    $content -match "flashTimer"
}

Test-Feature "Retry logic present" {
    $content = Get-Content "src\breez-service.cpp" -Raw
    $content -match "retryInitialization"
}

Write-Host ""
Write-Host "3. Documentation Tests" -ForegroundColor Yellow
Write-Host "----------------------" -ForegroundColor Yellow

Test-Feature "USER_GUIDE.md updated" {
    $content = Get-Content "USER_GUIDE.md" -Raw
    $content -match "Audio Notifications"
}

Test-Feature "RELEASE_CHECKLIST.md updated" {
    $content = Get-Content "RELEASE_CHECKLIST.md" -Raw
    $content -match "Flash effects" -and
    $content -match "audio"
}

Test-Feature "README.md exists" {
    Test-Path "README.md"
}

Write-Host ""
Write-Host "4. Installation Files" -ForegroundColor Yellow
Write-Host "---------------------" -ForegroundColor Yellow

Test-Feature "install.bat exists" {
    Test-Path "install.bat"
}

Test-Feature "install.sh exists" {
    Test-Path "install.sh"
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host " Test Results" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Passed: $testsPassed" -ForegroundColor Green
Write-Host "Failed: $testsFailed" -ForegroundColor Red
Write-Host ""

if ($testsFailed -eq 0) {
    Write-Host "All tests passed! ✓" -ForegroundColor Green
    exit 0
}
else {
    Write-Host "Some tests failed. Please review." -ForegroundColor Red
    exit 1
}

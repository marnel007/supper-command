# SuperShell Build Script
# This script builds the SuperShell application

Write-Host "Building SuperShell..." -ForegroundColor Green

# Set environment variables
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Clean previous builds
Write-Host "Cleaning previous builds..." -ForegroundColor Yellow
if (Test-Path "supershell.exe") {
    Remove-Item "supershell.exe" -Force
}

# Download dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Yellow
try {
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to download dependencies" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error downloading dependencies: $_" -ForegroundColor Red
    exit 1
}

# Build the application
Write-Host "Building application..." -ForegroundColor Yellow
try {
    go build -o supershell.exe ./cmd/supershell
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Build failed" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error building application: $_" -ForegroundColor Red
    exit 1
}

# Check if build was successful
if (Test-Path "supershell.exe") {
    $size = (Get-Item "supershell.exe").Length
    Write-Host "✅ Build successful!" -ForegroundColor Green
    Write-Host "Executable: supershell.exe" -ForegroundColor White
    Write-Host "Size: $([math]::Round($size/1MB, 2)) MB" -ForegroundColor White
    
    # Test the executable
    Write-Host "`nTesting executable..." -ForegroundColor Yellow
    try {
        $result = & .\supershell.exe --help 2>&1
        Write-Host "Executable test completed" -ForegroundColor Green
    } catch {
        Write-Host "Warning: Could not test executable: $_" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ Build failed - executable not found" -ForegroundColor Red
    exit 1
}

Write-Host "`nBuild completed successfully!" -ForegroundColor Green 
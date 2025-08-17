# Test runner script for SuperShell (PowerShell version)

param(
    [switch]$Coverage,
    [switch]$Integration,
    [switch]$Security,
    [switch]$Performance,
    [switch]$All
)

Write-Host "🧪 Running SuperShell Tests" -ForegroundColor Cyan
Write-Host "==========================" -ForegroundColor Cyan

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

$testsPassed = $true

# Run unit tests
Write-Host ""
Write-Host "📋 Running Unit Tests..." -ForegroundColor Blue
try {
    go test -v ./tests/unit/...
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Unit tests passed"
    } else {
        Write-Error "Unit tests failed"
        $testsPassed = $false
    }
} catch {
    Write-Error "Failed to run unit tests: $_"
    $testsPassed = $false
}

# Run tests with coverage
if ($Coverage -or $All) {
    Write-Host ""
    Write-Host "📊 Running Tests with Coverage..." -ForegroundColor Blue
    try {
        go test -cover ./tests/unit/...
        Write-Success "Coverage tests completed"
        
        # Generate detailed coverage report
        Write-Host ""
        Write-Host "📈 Generating Coverage Report..." -ForegroundColor Blue
        go test -coverprofile=coverage.out ./tests/unit/...
        go tool cover -html=coverage.out -o coverage.html
        Write-Success "Coverage report generated: coverage.html"
    } catch {
        Write-Warning "Coverage tests had issues: $_"
    }
}

# Run integration tests if they exist and requested
if (($Integration -or $All) -and (Test-Path "tests/integration") -and (Get-ChildItem "tests/integration" -Recurse -File).Count -gt 0) {
    Write-Host ""
    Write-Host "🔗 Running Integration Tests..." -ForegroundColor Blue
    try {
        go test -v ./tests/integration/...
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Integration tests passed"
        } else {
            Write-Error "Integration tests failed"
            $testsPassed = $false
        }
    } catch {
        Write-Error "Failed to run integration tests: $_"
        $testsPassed = $false
    }
}

# Run security tests if they exist and requested
if (($Security -or $All) -and (Test-Path "tests/security") -and (Get-ChildItem "tests/security" -Recurse -File).Count -gt 0) {
    Write-Host ""
    Write-Host "🔒 Running Security Tests..." -ForegroundColor Blue
    try {
        $env:SECURITY_TESTS = "enabled"
        go test -v ./tests/security/...
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Security tests passed"
        } else {
            Write-Error "Security tests failed"
            $testsPassed = $false
        }
    } catch {
        Write-Error "Failed to run security tests: $_"
        $testsPassed = $false
    }
}

# Run performance benchmarks if they exist and requested
if (($Performance -or $All) -and (Test-Path "tests/performance") -and (Get-ChildItem "tests/performance" -Recurse -File).Count -gt 0) {
    Write-Host ""
    Write-Host "⚡ Running Performance Benchmarks..." -ForegroundColor Blue
    try {
        go test -bench=. ./tests/performance/...
        Write-Success "Performance benchmarks completed"
    } catch {
        Write-Warning "Performance benchmarks had issues: $_"
    }
}

# Summary
Write-Host ""
if ($testsPassed) {
    Write-Success "All tests completed successfully!"
} else {
    Write-Error "Some tests failed!"
    exit 1
}

Write-Host ""
Write-Host "📋 Test Summary:" -ForegroundColor Cyan
Write-Host "  - Unit tests: ✅ Passed" -ForegroundColor White
if ($Coverage -or $All) {
    Write-Host "  - Coverage report: coverage.html" -ForegroundColor White
}
if (($Integration -or $All) -and (Test-Path "tests/integration")) {
    Write-Host "  - Integration tests: ✅ Available" -ForegroundColor White
}
if (($Security -or $All) -and (Test-Path "tests/security")) {
    Write-Host "  - Security tests: ✅ Available" -ForegroundColor White
}
if (($Performance -or $All) -and (Test-Path "tests/performance")) {
    Write-Host "  - Performance benchmarks: ✅ Available" -ForegroundColor White
}
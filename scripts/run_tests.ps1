# SuperShell Test Runner Script
# This script runs all tests for the SuperShell priority features implementation

param(
    [string]$TestType = "all",
    [switch]$Verbose,
    [switch]$Coverage,
    [switch]$Short
)

Write-Host "SuperShell Test Runner" -ForegroundColor Green
Write-Host "=====================" -ForegroundColor Green

# Set test environment variables
$env:GO_ENV = "test"

# Build test flags
$testFlags = @()
if ($Verbose) {
    $testFlags += "-v"
}
if ($Coverage) {
    $testFlags += "-cover"
    $testFlags += "-coverprofile=coverage.out"
}
if ($Short) {
    $testFlags += "-short"
}

# Function to run tests with proper error handling
function Run-Tests {
    param(
        [string]$TestPath,
        [string]$TestName
    )
    
    Write-Host "`nRunning $TestName..." -ForegroundColor Yellow
    Write-Host "Path: $TestPath" -ForegroundColor Gray
    
    $cmd = "go test $($testFlags -join ' ') $TestPath"
    Write-Host "Command: $cmd" -ForegroundColor Gray
    
    try {
        Invoke-Expression $cmd
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ $TestName passed" -ForegroundColor Green
        } else {
            Write-Host "❌ $TestName failed" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ $TestName failed with exception: $_" -ForegroundColor Red
        return $false
    }
    
    return $true
}

# Test execution based on type
$allPassed = $true

switch ($TestType.ToLower()) {
    "unit" {
        Write-Host "Running Unit Tests Only" -ForegroundColor Cyan
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/firewall" "Firewall Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/performance" "Performance Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/server" "Server Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/remote" "Remote Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/commands" "Commands Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/monitoring" "Monitoring Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/security" "Security Unit Tests")
    }
    
    "integration" {
        Write-Host "Running Integration Tests Only" -ForegroundColor Cyan
        $allPassed = $allPassed -and (Run-Tests "./tests/integration" "Cross-Platform Integration Tests")
    }
    
    "e2e" {
        Write-Host "Running End-to-End Tests Only" -ForegroundColor Cyan
        $allPassed = $allPassed -and (Run-Tests "./tests/e2e" "End-to-End Workflow Tests")
    }
    
    "all" {
        Write-Host "Running All Tests" -ForegroundColor Cyan
        
        # Unit Tests
        Write-Host "`n📋 UNIT TESTS" -ForegroundColor Magenta
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/firewall" "Firewall Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/performance" "Performance Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/server" "Server Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/remote" "Remote Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/commands" "Commands Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/monitoring" "Monitoring Unit Tests")
        $allPassed = $allPassed -and (Run-Tests "./tests/unit/security" "Security Unit Tests")
        
        # Integration Tests
        Write-Host "`n🔗 INTEGRATION TESTS" -ForegroundColor Magenta
        $allPassed = $allPassed -and (Run-Tests "./tests/integration" "Cross-Platform Integration Tests")
        
        # End-to-End Tests
        Write-Host "`n🎯 END-TO-END TESTS" -ForegroundColor Magenta
        $allPassed = $allPassed -and (Run-Tests "./tests/e2e" "End-to-End Workflow Tests")
    }
    
    default {
        Write-Host "Invalid test type: $TestType" -ForegroundColor Red
        Write-Host "Valid options: unit, integration, e2e, all" -ForegroundColor Yellow
        exit 1
    }
}

# Generate coverage report if requested
if ($Coverage -and (Test-Path "coverage.out")) {
    Write-Host "`nGenerating coverage report..." -ForegroundColor Yellow
    go tool cover -html=coverage.out -o coverage.html
    Write-Host "Coverage report generated: coverage.html" -ForegroundColor Green
}

# Final results
Write-Host "`n" + "="*50 -ForegroundColor Gray
if ($allPassed) {
    Write-Host "🎉 ALL TESTS PASSED!" -ForegroundColor Green
    Write-Host "SuperShell priority features are ready for deployment." -ForegroundColor Green
} else {
    Write-Host "❌ SOME TESTS FAILED!" -ForegroundColor Red
    Write-Host "Please review the test output above and fix any issues." -ForegroundColor Yellow
    exit 1
}

Write-Host "`nTest Summary:" -ForegroundColor Cyan
Write-Host "- Test Type: $TestType" -ForegroundColor White
Write-Host "- Verbose: $Verbose" -ForegroundColor White
Write-Host "- Coverage: $Coverage" -ForegroundColor White
Write-Host "- Short Mode: $Short" -ForegroundColor White

if ($Coverage -and (Test-Path "coverage.html")) {
    Write-Host "- Coverage Report: coverage.html" -ForegroundColor White
}

Write-Host "`nFeatures Tested:" -ForegroundColor Cyan
Write-Host "🔥 Firewall Management - Cross-platform firewall control" -ForegroundColor White
Write-Host "📊 Performance Analysis - System monitoring and optimization" -ForegroundColor White
Write-Host "🖥️  Server Management - Local server health and services" -ForegroundColor White
Write-Host "🌐 Remote Management - SSH-based remote operations" -ForegroundColor White
Write-Host "🏗️  Cluster Operations - Multi-server management" -ForegroundColor White
Write-Host "🔄 Config Synchronization - Configuration sync across servers" -ForegroundColor White

Write-Host "`nNext Steps:" -ForegroundColor Cyan
Write-Host "1. Review any test failures and fix issues" -ForegroundColor White
Write-Host "2. Run integration tests on target platforms" -ForegroundColor White
Write-Host "3. Perform manual testing of critical workflows" -ForegroundColor White
Write-Host "4. Update documentation with new features" -ForegroundColor White
Write-Host "5. Deploy to staging environment for validation" -ForegroundColor White
# Fix Go 1.14 compatibility issues
# Replace os.ReadFile and os.WriteFile with ioutil equivalents

Write-Host "Fixing Go 1.14 compatibility issues..." -ForegroundColor Green

# Files that need os.WriteFile -> ioutil.WriteFile
$writeFiles = @(
    "internal/managers/server/alerts.go",
    "internal/managers/performance/optimizer.go", 
    "internal/managers/firewall/windows.go",
    "internal/managers/firewall/linux.go",
    "internal/managers/performance/analyzer.go",
    "internal/core/builtins.go",
    "internal/commands/system/helphtml.go",
    "internal/commands/system/history_tracker.go",
    "internal/commands/system/smart_history.go",
    "internal/commands/system/bookmarks.go"
)

# Files that need os.ReadFile -> ioutil.ReadFile
$readFiles = @(
    "internal/managers/server/alerts.go",
    "internal/managers/performance/optimizer.go",
    "internal/managers/performance/analyzer.go", 
    "internal/managers/firewall/windows.go",
    "internal/managers/firewall/linux.go",
    "internal/commands/system/bookmarks.go",
    "internal/commands/system/history_tracker.go",
    "internal/commands/system/smart_history.go",
    "internal/commands/filesystem/cat.go"
)

# Fix os.WriteFile -> ioutil.WriteFile
foreach ($file in $writeFiles) {
    if (Test-Path $file) {
        Write-Host "Fixing $file (WriteFile)..." -ForegroundColor Yellow
        $content = Get-Content $file -Raw
        $content = $content -replace 'os\.WriteFile', 'ioutil.WriteFile'
        Set-Content $file $content -NoNewline
    }
}

# Fix os.ReadFile -> ioutil.ReadFile  
foreach ($file in $readFiles) {
    if (Test-Path $file) {
        Write-Host "Fixing $file (ReadFile)..." -ForegroundColor Yellow
        $content = Get-Content $file -Raw
        $content = $content -replace 'os\.ReadFile', 'ioutil.ReadFile'
        Set-Content $file $content -NoNewline
    }
}

# Add ioutil import to files that need it
$importFiles = @(
    "internal/managers/server/alerts.go",
    "internal/managers/performance/optimizer.go",
    "internal/managers/performance/analyzer.go",
    "internal/managers/firewall/windows.go", 
    "internal/managers/firewall/linux.go",
    "internal/commands/system/bookmarks.go",
    "internal/commands/system/history_tracker.go",
    "internal/commands/system/smart_history.go",
    "internal/commands/filesystem/cat.go"
)

foreach ($file in $importFiles) {
    if (Test-Path $file) {
        Write-Host "Adding ioutil import to $file..." -ForegroundColor Yellow
        $content = Get-Content $file -Raw
        if ($content -notmatch 'io/ioutil') {
            $content = $content -replace 'import \(', "import (`n`t`"io/ioutil`"`n"
            Set-Content $file $content -NoNewline
        }
    }
}

Write-Host "Go 1.14 compatibility fixes completed!" -ForegroundColor Green 
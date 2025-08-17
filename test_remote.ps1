# Test script for remote commands
Write-Host "=== Testing Remote Commands ==="

Write-Host "`n1. Adding remote servers..."
./supershell.exe -c "remote add web1 admin@localhost"
./supershell.exe -c "remote add db1 root@database.local"
./supershell.exe -c "remote add app1 deploy@app.example.com"

Write-Host "`n2. Testing other commands individually..."
Write-Host "Firewall status:"
./supershell.exe -c "firewall status"

Write-Host "`nPerformance analysis:"
./supershell.exe -c "perf analyze"

Write-Host "`nServer health:"
./supershell.exe -c "server health"

Write-Host "`n=== Remote Commands Test Complete ==="
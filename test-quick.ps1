# Quick Test Script for dsProxy
# Runs a minimal set of tests to verify the application is working

Write-Host "`n=== dsProxy Quick Test ===" -ForegroundColor Cyan

# Test write
Write-Host "`n1. Testing Write..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "http://localhost:8081/write" -Method POST `
    -Body '{"user_id":"quicktest","value":"success"}' `
    -ContentType "application/json"

# Test read
Write-Host "`n2. Testing Read..." -ForegroundColor Yellow
Invoke-RestMethod -Uri "http://localhost:8081/read?user_id=quicktest"

# Wait for batch
Write-Host "`n3. Waiting for batch flush (3s)..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

# Check database
Write-Host "`n4. Checking Database..." -ForegroundColor Yellow
docker exec pg psql -U postgres -d mydb -c "SELECT * FROM user_data WHERE user_id='quicktest';"

Write-Host "`n✅ Quick test complete!`n" -ForegroundColor Green

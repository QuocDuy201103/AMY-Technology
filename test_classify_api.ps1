# Test API Script
param(
    [string]$ServiceUrl = "https://cloud-inference-service-29216080826.us-central1.run.app",
    [string]$TestFile = "test_case.json"
)

Write-Host "`n=== Testing Cloud Inference API ===" -ForegroundColor Cyan
Write-Host "Service URL: $ServiceUrl" -ForegroundColor Yellow
Write-Host "Test file: $TestFile" -ForegroundColor Yellow

# Health check
Write-Host "`n[1] Health Check..." -ForegroundColor Cyan
try {
    $health = Invoke-RestMethod -Uri "$ServiceUrl/health" -Method Get
    Write-Host "✅ Health: OK" -ForegroundColor Green
} catch {
    Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Classify test
Write-Host "`n[2] Testing Classify..." -ForegroundColor Cyan

if (-not (Test-Path $TestFile)) {
    Write-Host "❌ Test file not found: $TestFile" -ForegroundColor Red
    exit 1
}

$body = Get-Content $TestFile -Raw
$emailCount = (($body | ConvertFrom-Json).emails).Count
Write-Host "Sending $emailCount emails..." -ForegroundColor Yellow

try {
    $response = Invoke-WebRequest -Uri "$ServiceUrl/classify" `
        -Method Post `
        -ContentType "application/json" `
        -Body $body
    
    $jsonResponse = $response.Content | ConvertFrom-Json
    
    Write-Host "`n✅ Success!" -ForegroundColor Green
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Emails processed: $($jsonResponse.results.Count)" -ForegroundColor Green
    
    # Count how many emails have exactly 1 label
    $singleLabelCount = ($jsonResponse.results | Where-Object { $_.labels.Count -eq 1 }).Count
    Write-Host "Emails with 1 label: $singleLabelCount / $($jsonResponse.results.Count)" -ForegroundColor Green
    
    Write-Host "`n=== Response ===" -ForegroundColor Cyan
    $jsonResponse | ConvertTo-Json -Depth 10
    
    Write-Host "`n=== Sample Results ===" -ForegroundColor Cyan
    foreach ($result in $jsonResponse.results) {
        Write-Host "`nEmail ID: $($result.id)" -ForegroundColor Yellow
        Write-Host "Labels:" -ForegroundColor Gray
        foreach ($label in $result.labels) {
            Write-Host "  - $($label.label) (score: $($label.score))" -ForegroundColor White
        }
    }
    
} catch {
    Write-Host "`n❌ Test failed: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Yellow
    }
    
    exit 1
}

Write-Host "`n🎉 All tests passed!" -ForegroundColor Green


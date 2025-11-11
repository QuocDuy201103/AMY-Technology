# Test All Endpoints - Comprehensive Test Suite
param(
    [string]$ServiceUrl = "https://cloud-inference-service-29216080826.us-central1.run.app"
)

$ErrorActionPreference = "Stop"
$testsPassed = 0
$testsFailed = 0

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Cloud Inference API - Full Test Suite" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Service URL: $ServiceUrl" -ForegroundColor Yellow
Write-Host ""

# Test 1: Health Check
Write-Host "[Test 1/4] Health Check..." -ForegroundColor Cyan
try {
    $health = Invoke-RestMethod -Uri "$ServiceUrl/health" -Method Get
    Write-Host "  PASSED - Health check OK" -ForegroundColor Green
    $testsPassed++
} catch {
    Write-Host "  FAILED - Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    $testsFailed++
}

Write-Host ""

# Test 2: Classify (Batch)
Write-Host "[Test 2/4] Classify (Batch Processing)..." -ForegroundColor Cyan
try {
    $classifyData = @{
        emails = @(
            @{
                id = "test-1"
                content = "URGENT: Please approve the budget by EOD!"
            },
            @{
                id = "test-2"
                content = "Hey! Want to grab lunch tomorrow?"
            }
        )
    } | ConvertTo-Json -Depth 10
    
    $response = Invoke-WebRequest `
        -Uri "$ServiceUrl/classify" `
        -Method Post `
        -ContentType "application/json" `
        -Body $classifyData
    
    $result = $response.Content | ConvertFrom-Json
    
    if ($result.results.Count -eq 2) {
        $singleLabel = ($result.results | Where-Object { $_.labels.Count -eq 1 }).Count
        if ($singleLabel -eq 2) {
            Write-Host "  PASSED - Classified 2 emails, each with 1 label" -ForegroundColor Green
            Write-Host "    Email 1: $($result.results[0].labels[0].label) (score: $($result.results[0].labels[0].score))" -ForegroundColor Gray
            Write-Host "    Email 2: $($result.results[1].labels[0].label) (score: $($result.results[1].labels[0].score))" -ForegroundColor Gray
            $testsPassed++
        } else {
            Write-Host "  FAILED - Not all emails have exactly 1 label" -ForegroundColor Red
            $testsFailed++
        }
    } else {
        Write-Host "  FAILED - Expected 2 results, got $($result.results.Count)" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "  FAILED - Classify test failed: $($_.Exception.Message)" -ForegroundColor Red
    $testsFailed++
}

Write-Host ""

# Test 3: Summarize
Write-Host "[Test 3/4] Summarize..." -ForegroundColor Cyan
try {
    $emailContent = @"
<html><body>
<p>Hi Team,</p>
<p>Following up on yesterday's meeting about Q4 budget. Key points:</p>
<ul>
<li>Increase marketing budget by 15%</li>
<li>Reduce operational costs</li>
<li>Allocate funds for training</li>
</ul>
<p>Please review by Friday. Board meeting next week.</p>
<p>Best, John</p>
</body></html>
"@
    
    $response = Invoke-WebRequest `
        -Uri "$ServiceUrl/summarize" `
        -Method Post `
        -ContentType "text/plain" `
        -Body $emailContent
    
    $result = $response.Content | ConvertFrom-Json
    
    if ($result.summary -and $result.summary.Length -gt 10) {
        Write-Host "  PASSED - Generated summary" -ForegroundColor Green
        Write-Host "    Summary: $($result.summary.Substring(0, [Math]::Min(80, $result.summary.Length)))..." -ForegroundColor Gray
        $testsPassed++
    } else {
        Write-Host "  FAILED - Summary is empty or too short" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "  FAILED - Summarize test failed: $($_.Exception.Message)" -ForegroundColor Red
    $testsFailed++
}

Write-Host ""

# Test 4: Draft Reply
Write-Host "[Test 4/4] Draft Reply..." -ForegroundColor Cyan
try {
    $emailContent = @"
<html><body>
<p>Hi Sarah,</p>
<p>Our team has encountered technical issues that might delay the project by 3-5 days. 
Would it be possible to schedule a call tomorrow after 2 PM to discuss?</p>
<p>Best, Michael</p>
</body></html>
"@
    
    $response = Invoke-WebRequest `
        -Uri "$ServiceUrl/draft" `
        -Method Post `
        -ContentType "text/plain" `
        -Body $emailContent
    
    $result = $response.Content | ConvertFrom-Json
    
    if ($result.draft -and $result.draft.Length -gt 10) {
        Write-Host "  PASSED - Generated draft reply" -ForegroundColor Green
        Write-Host "    Draft: $($result.draft.Substring(0, [Math]::Min(80, $result.draft.Length)))..." -ForegroundColor Gray
        $testsPassed++
    } else {
        Write-Host "  FAILED - Draft is empty or too short" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "  FAILED - Draft test failed: $($_.Exception.Message)" -ForegroundColor Red
    $testsFailed++
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "           Test Summary" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Tests Passed: $testsPassed / $($testsPassed + $testsFailed)" -ForegroundColor $(if ($testsPassed -eq 4) { "Green" } else { "Yellow" })
Write-Host "Tests Failed: $testsFailed / $($testsPassed + $testsFailed)" -ForegroundColor $(if ($testsFailed -eq 0) { "Green" } else { "Red" })
Write-Host ""

if ($testsFailed -eq 0) {
    Write-Host "ALL TESTS PASSED!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "SOME TESTS FAILED!" -ForegroundColor Red
    exit 1
}


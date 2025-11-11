# Test Summarize Endpoint
param(
    [string]$ServiceUrl = "https://cloud-inference-service-29216080826.us-central1.run.app"
)

Write-Host ""
Write-Host "=== Testing Summarize Endpoint ===" -ForegroundColor Cyan
Write-Host "Service URL: $ServiceUrl" -ForegroundColor Yellow

# Test data - email content
$emailContent = @"
<html>
<body>
<p>Hi Team,</p>

<p>I wanted to follow up on our discussion from yesterday's meeting regarding the Q4 budget allocation. 
After reviewing the numbers with the finance team, I believe we need to make some adjustments to our 
initial proposal.</p>

<p>Here are the key points:</p>
<ul>
<li>Increase marketing budget by 15% to support the new product launch</li>
<li>Reduce operational costs by optimizing our cloud infrastructure</li>
<li>Allocate additional funds for employee training and development</li>
</ul>

<p>Please review the attached updated proposal and let me know your thoughts by Friday, November 15th. 
We need to finalize this before the board meeting next week.</p>

<p>Also, don't forget about the team lunch on Thursday at 12:30 PM.</p>

<p>Best regards,<br>
John Smith<br>
Senior Manager</p>
</body>
</html>
"@

Write-Host ""
Write-Host "[1] Sending email for summarization..." -ForegroundColor Cyan
Write-Host "Email length: $($emailContent.Length) characters" -ForegroundColor Gray

try {
    $response = Invoke-WebRequest `
        -Uri "$ServiceUrl/summarize" `
        -Method Post `
        -ContentType "text/plain" `
        -Body $emailContent
    
    $result = $response.Content | ConvertFrom-Json
    
    Write-Host ""
    Write-Host "SUCCESS!" -ForegroundColor Green
    Write-Host "Status Code: $($response.StatusCode)" -ForegroundColor Green
    
    Write-Host ""
    Write-Host "=== Summary ===" -ForegroundColor Cyan
    Write-Host $result.summary -ForegroundColor White
    
    Write-Host ""
    Write-Host "=== Full JSON Response ===" -ForegroundColor Cyan
    $result | ConvertTo-Json -Depth 10
    
    Write-Host ""
    Write-Host "Test passed!" -ForegroundColor Green
    
} catch {
    Write-Host ""
    Write-Host "Test failed: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response: $responseBody" -ForegroundColor Yellow
    }
    
    exit 1
}


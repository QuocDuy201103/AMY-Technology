# Test Draft Reply Endpoint
param(
    [string]$ServiceUrl = "https://cloud-inference-service-29216080826.us-central1.run.app"
)

Write-Host ""
Write-Host "=== Testing Draft Reply Endpoint ===" -ForegroundColor Cyan
Write-Host "Service URL: $ServiceUrl" -ForegroundColor Yellow

# Test data - email content to reply to
$emailContent = @"
<html>
<body>
<p>Hi Sarah,</p>

<p>I hope this email finds you well. I wanted to reach out regarding the upcoming project deadline 
for the client presentation.</p>

<p>Unfortunately, our team has encountered some technical issues with the data integration that 
might delay our original timeline by 3-5 business days. We're working around the clock to resolve 
these issues, but I wanted to give you a heads up as soon as possible.</p>

<p>Would it be possible to schedule a quick call tomorrow afternoon to discuss potential solutions 
and revised timelines? I'm available anytime after 2 PM.</p>

<p>I apologize for any inconvenience this may cause.</p>

<p>Best regards,<br>
Michael Chen<br>
Project Lead</p>
</body>
</html>
"@

Write-Host ""
Write-Host "[1] Sending email for draft reply..." -ForegroundColor Cyan
Write-Host "Email length: $($emailContent.Length) characters" -ForegroundColor Gray

try {
    $response = Invoke-WebRequest `
        -Uri "$ServiceUrl/draft" `
        -Method Post `
        -ContentType "text/plain" `
        -Body $emailContent
    
    $result = $response.Content | ConvertFrom-Json
    
    Write-Host ""
    Write-Host "SUCCESS!" -ForegroundColor Green
    Write-Host "Status Code: $($response.StatusCode)" -ForegroundColor Green
    
    Write-Host ""
    Write-Host "=== Draft Reply ===" -ForegroundColor Cyan
    Write-Host $result.draft -ForegroundColor White
    
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


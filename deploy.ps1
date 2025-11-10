# Deploy Script - Build & Deploy to Cloud Run
$PROJECT_ID = "siftly-backend-dev"
$REGION = "us-central1"
$SERVICE_NAME = "cloud-inference-service"
$IMAGE_NAME = "cloud-inference"

Write-Host "`n=== Cloud Inference Deployment ===" -ForegroundColor Cyan
Write-Host "Project: $PROJECT_ID" -ForegroundColor Yellow
Write-Host "Region: $REGION" -ForegroundColor Yellow
Write-Host "Service: $SERVICE_NAME" -ForegroundColor Yellow

# Confirm
Write-Host "`nThis will build and deploy the service. Continue? (Y/N)" -ForegroundColor Yellow
$confirm = Read-Host
if ($confirm -ne "Y" -and $confirm -ne "y") {
    Write-Host "Deployment cancelled." -ForegroundColor Red
    exit 0
}

# Step 1: Build & Push Image
Write-Host "`n=== Step 1: Building Image ===" -ForegroundColor Cyan
Write-Host "Running: gcloud builds submit..." -ForegroundColor Gray

gcloud builds submit --tag gcr.io/$PROJECT_ID/${IMAGE_NAME}:latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ Build completed successfully!" -ForegroundColor Green

# Step 2: Deploy to Cloud Run
Write-Host "`n=== Step 2: Deploying to Cloud Run ===" -ForegroundColor Cyan
Write-Host "Running: gcloud run deploy..." -ForegroundColor Gray

gcloud run deploy $SERVICE_NAME `
    --image gcr.io/$PROJECT_ID/${IMAGE_NAME}:latest `
    --region $REGION `
    --platform managed `
    --allow-unauthenticated `
    --memory 512Mi `
    --cpu 1 `
    --timeout 300 `
    --max-instances 10 `
    --set-env-vars DEEPSEEK_API_URL=https://api.deepseek.com `
    --set-secrets DEEPSEEK_API_KEY=deepseek-api-key:latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Deploy failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n✅ Deploy completed successfully!" -ForegroundColor Green

# Step 3: Get Service URL
Write-Host "`n=== Step 3: Getting Service Info ===" -ForegroundColor Cyan

$SERVICE_URL = gcloud run services describe $SERVICE_NAME `
    --region $REGION `
    --format="value(status.url)"

if (-not $SERVICE_URL) {
    Write-Host "❌ Failed to get service URL" -ForegroundColor Red
    exit 1
}

Write-Host "Service URL: $SERVICE_URL" -ForegroundColor Green

# Step 4: Test Deployment
Write-Host "`n=== Step 4: Testing Deployment ===" -ForegroundColor Cyan

Write-Host "Testing health check..." -ForegroundColor Gray
try {
    $health = Invoke-RestMethod -Uri "$SERVICE_URL/health" -Method Get
    Write-Host "✅ Health check: OK" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Health check failed: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host "`nTesting classify endpoint..." -ForegroundColor Gray
if (Test-Path "test_case.json") {
    try {
        $body = Get-Content test_case.json -Raw
        $response = Invoke-WebRequest -Uri "$SERVICE_URL/classify" `
            -Method Post `
            -ContentType "application/json" `
            -Body $body
        
        $jsonResponse = $response.Content | ConvertFrom-Json
        
        Write-Host "✅ Classify test: OK" -ForegroundColor Green
        Write-Host "   Emails processed: $($jsonResponse.results.Count)" -ForegroundColor Green
        $totalLabels = ($jsonResponse.results | ForEach-Object { $_.labels.Count } | Measure-Object -Sum).Sum
        Write-Host "   Total labels: $totalLabels" -ForegroundColor Green
        
    } catch {
        Write-Host "⚠️  Classify test failed: $($_.Exception.Message)" -ForegroundColor Yellow
    }
} else {
    Write-Host "⚠️  test_case.json not found, skipping classify test" -ForegroundColor Yellow
}

# Summary
Write-Host "`n=== Deployment Summary ===" -ForegroundColor Cyan
Write-Host "✅ Build: Success" -ForegroundColor Green
Write-Host "✅ Deploy: Success" -ForegroundColor Green
Write-Host "✅ Service URL: $SERVICE_URL" -ForegroundColor Green

Write-Host "`n=== Next Steps ===" -ForegroundColor Cyan
Write-Host "Test API with: .\test_api.ps1" -ForegroundColor White
Write-Host "View logs: gcloud run services logs read $SERVICE_NAME --region $REGION --limit 50" -ForegroundColor White

Write-Host "`n🎉 Deployment completed successfully!" -ForegroundColor Green


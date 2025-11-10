# Quick Start Guide - Chạy Lại API

## Test API (Ngay Bây Giờ)

API đã deploy thành công, chỉ cần test:

### Test với PowerShell

```powershell
# Test với 5 emails (batch)
$body = Get-Content test_case.json -Raw
$response = Invoke-WebRequest -Uri "https://cloud-inference-service-29216080826.us-central1.run.app/classify" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body

$jsonResponse = $response.Content | ConvertFrom-Json
$jsonResponse | ConvertTo-Json -Depth 10
```

### Test Health Check

```powershell
Invoke-RestMethod -Uri "https://cloud-inference-service-29216080826.us-central1.run.app/health" -Method Get
```

---

## Deploy Lại (Khi Có Thay Đổi Code)

Nếu bạn thay đổi code trong `main.go` hoặc `deepseek_client.go`:

### Option 1: Deploy Bằng Gcloud (Nhanh)

```powershell
# Bước 1: Build & push image mới
gcloud builds submit --tag gcr.io/siftly-backend-dev/cloud-inference:latest

# Bước 2: Deploy lên Cloud Run
gcloud run deploy cloud-inference-service `
    --image gcr.io/siftly-backend-dev/cloud-inference:latest `
    --region us-central1 `
    --platform managed `
    --allow-unauthenticated `
    --memory 512Mi `
    --cpu 1 `
    --timeout 300 `
    --set-env-vars DEEPSEEK_API_URL=https://api.deepseek.com `
    --set-secrets DEEPSEEK_API_KEY=deepseek-api-key:latest

# Bước 3: Test
$body = Get-Content test_case.json -Raw
$response = Invoke-WebRequest -Uri "https://cloud-inference-service-29216080826.us-central1.run.app/classify" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
$response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
```

### Option 2: Deploy Bằng Pulumi

```powershell
# Bước 1: Build & push image mới
gcloud builds submit --tag gcr.io/siftly-backend-dev/cloud-inference:latest

# Bước 2: Deploy với Pulumi
pulumi up

# Bước 3: Test
$SERVICE_URL = pulumi stack output serviceUrl
$body = Get-Content test_case.json -Raw
$response = Invoke-WebRequest -Uri "$SERVICE_URL/classify" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body
$response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
```

---

## Chạy Local (Test Trên Máy)

Nếu muốn test trên máy local trước khi deploy:

### Bước 1: Set Environment Variables

```powershell
$env:DEEPSEEK_API_KEY = "your-api-key-here"
$env:DEEPSEEK_API_URL = "https://api.deepseek.com"
```

### Bước 2: Build & Run

```powershell
# Build
go build -o server.exe

# Run
.\server.exe
```

### Bước 3: Test Local

```powershell
$body = Get-Content test_case.json -Raw
$response = Invoke-WebRequest -Uri "http://localhost:8080/classify" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body

$response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
```

---

## Script Tự Động (All-in-One)

### Deploy Script

Lưu vào file `deploy.ps1`:

```powershell
# deploy.ps1
$PROJECT_ID = "siftly-backend-dev"
$REGION = "us-central1"
$SERVICE_NAME = "cloud-inference-service"

Write-Host "`n=== Step 1: Build Image ===" -ForegroundColor Cyan
gcloud builds submit --tag gcr.io/$PROJECT_ID/cloud-inference:latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== Step 2: Deploy to Cloud Run ===" -ForegroundColor Cyan
gcloud run deploy $SERVICE_NAME `
    --image gcr.io/$PROJECT_ID/cloud-inference:latest `
    --region $REGION `
    --platform managed `
    --allow-unauthenticated `
    --memory 512Mi `
    --cpu 1 `
    --timeout 300 `
    --set-env-vars DEEPSEEK_API_URL=https://api.deepseek.com `
    --set-secrets DEEPSEEK_API_KEY=deepseek-api-key:latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Deploy failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n=== Step 3: Get Service URL ===" -ForegroundColor Cyan
$SERVICE_URL = gcloud run services describe $SERVICE_NAME `
    --region $REGION `
    --format="value(status.url)"
Write-Host "Service URL: $SERVICE_URL" -ForegroundColor Green

Write-Host "`n=== Step 4: Test ===" -ForegroundColor Cyan
$body = Get-Content test_case.json -Raw
$response = Invoke-WebRequest -Uri "$SERVICE_URL/classify" `
    -Method Post `
    -ContentType "application/json" `
    -Body $body

$jsonResponse = $response.Content | ConvertFrom-Json
Write-Host "`n✅ Deployment Success!" -ForegroundColor Green
Write-Host "Emails processed: $($jsonResponse.results.Count)" -ForegroundColor Green
$jsonResponse | ConvertTo-Json -Depth 10
```

**Chạy script:**

```powershell
.\deploy.ps1
```

### Test Script

Lưu vào file `test_api.ps1`:

```powershell
# test_api.ps1
param(
    [string]$ServiceUrl = "https://cloud-inference-service-29216080826.us-central1.run.app",
    [string]$TestFile = "test_case.json"
)

Write-Host "`n=== Testing API ===" -ForegroundColor Cyan
Write-Host "Service URL: $ServiceUrl" -ForegroundColor Yellow
Write-Host "Test file: $TestFile" -ForegroundColor Yellow

# Health check
Write-Host "`n[1] Health Check..." -ForegroundColor Cyan
try {
    $health = Invoke-RestMethod -Uri "$ServiceUrl/health" -Method Get
    Write-Host "✅ Health: $($health | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "❌ Health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Classify test
Write-Host "`n[2] Testing Classify..." -ForegroundColor Cyan
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
    
    $totalLabels = ($jsonResponse.results | ForEach-Object { $_.labels.Count } | Measure-Object -Sum).Sum
    Write-Host "Total labels: $totalLabels" -ForegroundColor Green
    
    Write-Host "`n=== Response ===" -ForegroundColor Cyan
    $jsonResponse | ConvertTo-Json -Depth 10
    
} catch {
    Write-Host "❌ Test failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
```

**Chạy test:**

```powershell
# Test với default URL
.\test_api.ps1

# Test với custom URL
.\test_api.ps1 -ServiceUrl "https://your-service-url.run.app"
```

---

## Tóm Tắt Lệnh Thường Dùng

### Chỉ Test (API Đã Deploy)

```powershell
$body = Get-Content test_case.json -Raw
$response = Invoke-WebRequest -Uri "https://cloud-inference-service-29216080826.us-central1.run.app/classify" -Method Post -ContentType "application/json" -Body $body
$response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10
```

### Deploy Lại (Khi Có Thay Đổi Code)

```powershell
gcloud builds submit --tag gcr.io/siftly-backend-dev/cloud-inference:latest
gcloud run deploy cloud-inference-service --image gcr.io/siftly-backend-dev/cloud-inference:latest --region us-central1 --platform managed --allow-unauthenticated --set-env-vars DEEPSEEK_API_URL=https://api.deepseek.com --set-secrets DEEPSEEK_API_KEY=deepseek-api-key:latest
```

### Xem Logs

```powershell
gcloud run services logs read cloud-inference-service --region us-central1 --limit 50
```

### Kiểm Tra Service Info

```powershell
gcloud run services describe cloud-inference-service --region us-central1
```

---

## Troubleshooting

### Lỗi 500 "Failed to classify email"

**Nguyên nhân:** Code cũ đang chạy trên Cloud Run

**Giải pháp:** Build & deploy lại (xem phần "Deploy Lại" ở trên)

### Labels Rỗng

**Nguyên nhân:** Dùng `Invoke-RestMethod` thay vì `Invoke-WebRequest`

**Giải pháp:** Luôn dùng `Invoke-WebRequest` với manual `ConvertFrom-Json`:

```powershell
# ❌ SAI
$response = Invoke-RestMethod -Uri "..." -Method Post -Body $body

# ✅ ĐÚNG
$response = Invoke-WebRequest -Uri "..." -Method Post -Body $body
$jsonResponse = $response.Content | ConvertFrom-Json
```

### Build Failed

**Nguyên nhân:** Docker daemon không chạy hoặc chưa login gcloud

**Giải pháp:**

```powershell
# Login lại
gcloud auth login

# Set project
gcloud config set project siftly-backend-dev
```

---

## Next Steps

1. **Test API ngay:** Chạy test script để verify
2. **Tích hợp vào app:** Sử dụng API endpoint trong ứng dụng của bạn
3. **Monitor:** Theo dõi logs và performance
4. **Scale:** Điều chỉnh `--memory`, `--cpu`, `--max-instances` nếu cần

---

**Current Service URL:**
```
https://cloud-inference-service-29216080826.us-central1.run.app
```

**Status:** ✅ Ready to use


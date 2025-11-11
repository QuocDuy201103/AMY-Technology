# Test Scripts Guide

## Tổng Quan

Project có 4 test scripts để test các endpoints của Cloud Inference API:

1. **`test_api.ps1`** - Test Classify endpoint (batch processing)
2. **`test_summarize.ps1`** - Test Summarize endpoint
3. **`test_draft.ps1`** - Test Draft Reply endpoint
4. **`test_all_endpoints.ps1`** - Test tất cả endpoints (khuyến nghị)

---

## 1. test_api.ps1

### Mục Đích
Test email classification với batch processing (1-100 emails).

### Cách Chạy

```powershell
.\test_api.ps1
```

### Custom URL

```powershell
.\test_api.ps1 -ServiceUrl "https://your-service-url.run.app"
```

### Kết Quả Mong Đợi
- ✅ Health check thành công
- ✅ Classify 5 emails từ `test_case.json`
- ✅ Mỗi email có **chỉ 1 label** với score cao nhất
- ✅ Response format: `{"results": [{"id": "...", "labels": [...]}]}`

### Example Output

```
=== Testing Cloud Inference API ===
Service URL: https://cloud-inference-service-29216080826.us-central1.run.app

[1] Health Check...
✅ Health: OK

[2] Testing Classify...
Sending 5 emails...

✅ Success!
Status: 200
Emails processed: 5
Emails with 1 label: 5 / 5

=== Sample Results ===
Email ID: email-001
Labels:
  - follow_up (score: 0.95)

Email ID: email-002
Labels:
  - urgent (score: 0.95)
...
```

---

## 2. test_summarize.ps1

### Mục Đích
Test email summarization endpoint.

### Cách Chạy

```powershell
.\test_summarize.ps1
```

### Custom URL

```powershell
.\test_summarize.ps1 -ServiceUrl "https://your-service-url.run.app"
```

### Kết Quả Mong Đợi
- ✅ Tóm tắt email thành công
- ✅ Response format: `{"summary": "..."}`
- ✅ Summary ngắn gọn và chính xác

### Example Output

```
=== Testing Summarize Endpoint ===

[1] Sending email for summarization...
Email length: 864 characters

✅ SUCCESS!

=== Summary ===
John Smith proposes Q4 budget adjustments: increase marketing by 15% 
for new product launch, reduce operational costs via cloud optimization, 
and allocate more funds for employee training.

✅ Test passed!
```

### Request Format

```powershell
# Email content as plain text (HTML allowed)
$emailContent = @"
<html><body>
<p>Your email content here...</p>
</body></html>
"@

$response = Invoke-WebRequest `
    -Uri "$ServiceUrl/summarize" `
    -Method Post `
    -ContentType "text/plain" `
    -Body $emailContent
```

---

## 3. test_draft.ps1

### Mục Đích
Test email draft reply generation endpoint.

### Cách Chạy

```powershell
.\test_draft.ps1
```

### Custom URL

```powershell
.\test_draft.ps1 -ServiceUrl "https://your-service-url.run.app"
```

### Kết Quả Mong Đợi
- ✅ Tạo draft reply thành công
- ✅ Response format: `{"draft": "..."}`
- ✅ Draft lịch sự và phù hợp với ngữ cảnh

### Example Output

```
=== Testing Draft Reply Endpoint ===

[1] Sending email for draft reply...
Email length: 748 characters

✅ SUCCESS!

=== Draft Reply ===
<p>Hi Michael,</p>
<p>Thank you for letting me know about the delay. I appreciate 
the proactive update.</p>
<p>I'm available for a call tomorrow after 2 PM as well...</p>
<p>Best regards,<br>Sarah</p>

✅ Test passed!
```

### Request Format

```powershell
# Email content to reply to (plain text, HTML allowed)
$emailContent = @"
<html><body>
<p>Email you want to reply to...</p>
</body></html>
"@

$response = Invoke-WebRequest `
    -Uri "$ServiceUrl/draft" `
    -Method Post `
    -ContentType "text/plain" `
    -Body $emailContent
```

---

## 4. test_all_endpoints.ps1 ⭐ (Khuyến Nghị)

### Mục Đích
Test **tất cả endpoints** trong một lần chạy.

### Cách Chạy

```powershell
.\test_all_endpoints.ps1
```

### Custom URL

```powershell
.\test_all_endpoints.ps1 -ServiceUrl "https://your-service-url.run.app"
```

### Tests Được Thực Hiện

1. **Health Check** - Kiểm tra service đang chạy
2. **Classify** - Test batch classification với 2 emails
3. **Summarize** - Test email summarization
4. **Draft Reply** - Test draft reply generation

### Kết Quả Mong Đợi
- ✅ Tất cả 4 tests pass
- ✅ Exit code: 0
- ❌ Nếu có test fail: Exit code: 1

### Example Output

```
======================================
  Cloud Inference API - Full Test Suite
======================================
Service URL: https://cloud-inference-service-29216080826.us-central1.run.app

[Test 1/4] Health Check...
  ✅ PASSED - Health check OK

[Test 2/4] Classify (Batch Processing)...
  ✅ PASSED - Classified 2 emails, each with 1 label
    Email 1: urgent (score: 0.95)
    Email 2: personal (score: 0.95)

[Test 3/4] Summarize...
  ✅ PASSED - Generated summary

[Test 4/4] Draft Reply...
  ✅ PASSED - Generated draft reply

======================================
           Test Summary
======================================
Tests Passed: 4 / 4
Tests Failed: 0 / 4

✅ ALL TESTS PASSED!
```

---

## So Sánh Scripts

| Script | Endpoints | Use Case | Test Data |
|--------|-----------|----------|-----------|
| `test_api.ps1` | Classify | Test batch classification | 5 emails từ `test_case.json` |
| `test_summarize.ps1` | Summarize | Test email summarization | 1 email trong script |
| `test_draft.ps1` | Draft | Test draft reply | 1 email trong script |
| `test_all_endpoints.ps1` | All | Comprehensive testing | Tất cả (2 emails classify + 1 summarize + 1 draft) |

---

## Khi Nào Dùng Script Nào?

### Dùng `test_all_endpoints.ps1` khi:
- ✅ Sau khi deploy code mới
- ✅ Muốn verify toàn bộ API
- ✅ Chạy regression test
- ✅ CI/CD pipeline

### Dùng `test_api.ps1` khi:
- ✅ Chỉ test classify endpoint
- ✅ Test với nhiều emails (từ file JSON)
- ✅ Debug classification logic

### Dùng `test_summarize.ps1` khi:
- ✅ Chỉ test summarize endpoint
- ✅ Debug summarization logic
- ✅ Test với custom email content

### Dùng `test_draft.ps1` khi:
- ✅ Chỉ test draft reply endpoint
- ✅ Debug draft generation logic
- ✅ Test với custom email scenarios

---

## Tips & Tricks

### 1. Chạy Tất Cả Tests Liên Tiếp

```powershell
# Chạy tất cả tests
.\test_all_endpoints.ps1

# Hoặc chạy từng test riêng
.\test_api.ps1
.\test_summarize.ps1
.\test_draft.ps1
```

### 2. Test với Custom Service URL

```powershell
$URL = "https://your-custom-url.run.app"
.\test_all_endpoints.ps1 -ServiceUrl $URL
```

### 3. Capture Output

```powershell
# Save output to file
.\test_all_endpoints.ps1 | Tee-Object -FilePath "test-results.txt"
```

### 4. Check Exit Code

```powershell
.\test_all_endpoints.ps1
if ($LASTEXITCODE -eq 0) {
    Write-Host "All tests passed!" -ForegroundColor Green
} else {
    Write-Host "Some tests failed!" -ForegroundColor Red
}
```

---

## Troubleshooting

### Error: "Cannot find path test_case.json"

**Solution:** Đảm bảo file `test_case.json` tồn tại trong thư mục hiện tại.

```powershell
# Check if file exists
Test-Path test_case.json
```

### Error: "Connection refused" / "Service unavailable"

**Solution:** 
1. Kiểm tra service URL đúng chưa
2. Verify service đang chạy: `gcloud run services describe cloud-inference-service --region us-central1`
3. Check logs: `gcloud run services logs read cloud-inference-service --region us-central1 --limit 50`

### Error: "Test failed" - Empty response

**Solution:**
1. Check DEEPSEEK_API_KEY đã set chưa
2. Verify service có đủ quyền truy cập Secret Manager
3. Check logs để xem lỗi chi tiết

---

## Test Data Files

### test_case.json

File chứa 5 test emails cho batch classification:

```json
{
  "emails": [
    {"id": "email-001", "content": "Follow-up email..."},
    {"id": "email-002", "content": "Urgent action required..."},
    {"id": "email-003", "content": "Spam/scam email..."},
    {"id": "email-004", "content": "Personal/social email..."},
    {"id": "email-005", "content": "Meeting reminder..."}
  ]
}
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Test API Endpoints
  run: |
    pwsh -File test_all_endpoints.ps1 -ServiceUrl ${{ secrets.SERVICE_URL }}
  
- name: Check Test Results
  if: failure()
  run: |
    echo "API tests failed!"
    exit 1
```

### Cloud Build Example

```yaml
steps:
  - name: 'gcr.io/cloud-builders/gcloud'
    entrypoint: 'pwsh'
    args: ['test_all_endpoints.ps1']
```

---

## Service URL

**Current Production URL:**
```
https://cloud-inference-service-29216080826.us-central1.run.app
```

**Status:** ✅ Live and Working

---

## Next Steps

1. ✅ Chạy `test_all_endpoints.ps1` để verify API
2. ✅ Integrate tests vào CI/CD pipeline
3. ✅ Customize test data cho use cases cụ thể
4. ✅ Add monitoring và alerting dựa trên test results

---

**Last Updated:** $(Get-Date -Format 'yyyy-MM-dd HH:mm')


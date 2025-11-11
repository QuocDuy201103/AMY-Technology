# Test Scripts - Quick Reference

## 🚀 Quick Start

### Test Tất Cả Endpoints (Khuyến Nghị)

```powershell
.\test_all_endpoints.ps1
```

**Kết quả:** Test 4 endpoints (Health, Classify, Summarize, Draft) - tất cả đều PASSED ✅

---

## 📋 Available Scripts

| Script | Endpoint | Mô Tả | Thời Gian |
|--------|----------|-------|-----------|
| `test_all_endpoints.ps1` | ALL | Test tất cả endpoints | ~15-20s |
| `test_api.ps1` | `/classify` | Test batch classification (5 emails) | ~10s |
| `test_summarize.ps1` | `/summarize` | Test email summarization | ~5s |
| `test_draft.ps1` | `/draft` | Test draft reply generation | ~5s |

---

## 💡 Cách Sử Dụng

### 1. Test Tất Cả (All-in-One)

```powershell
.\test_all_endpoints.ps1
```

**Output:**
```
[Test 1/4] Health Check... ✅ PASSED
[Test 2/4] Classify...     ✅ PASSED (2 emails, each with 1 label)
[Test 3/4] Summarize...    ✅ PASSED
[Test 4/4] Draft Reply...  ✅ PASSED

Tests Passed: 4 / 4
ALL TESTS PASSED! ✅
```

### 2. Test Classify (Batch)

```powershell
.\test_api.ps1
```

**Output:**
- Test với 5 emails từ `test_case.json`
- Mỗi email trả về **1 label duy nhất** với score cao nhất
- Format: `{"results": [{"id": "...", "labels": [...]}]}`

### 3. Test Summarize

```powershell
.\test_summarize.ps1
```

**Output:**
- Tóm tắt nội dung email
- Format: `{"summary": "..."}`

### 4. Test Draft Reply

```powershell
.\test_draft.ps1
```

**Output:**
- Tạo draft reply cho email
- Format: `{"draft": "..."}`

---

## 🎯 Endpoints

### 1. POST /classify

**Request:**
```json
{
  "emails": [
    {"id": "email-1", "content": "Email content..."},
    {"id": "email-2", "content": "Another email..."}
  ]
}
```

**Response:** *(Mỗi email chỉ 1 label)*
```json
{
  "results": [
    {"id": "email-1", "labels": [{"label": "urgent", "score": 0.95}]},
    {"id": "email-2", "labels": [{"label": "personal", "score": 0.95}]}
  ]
}
```

### 2. POST /summarize

**Request:** Plain text email content

**Response:**
```json
{
  "summary": "Brief summary of the email content..."
}
```

### 3. POST /draft

**Request:** Plain text email content to reply to

**Response:**
```json
{
  "draft": "<p>Draft reply content...</p>"
}
```

### 4. GET /health

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-11-11T10:30:00Z"
}
```

---

## 📊 Test Results Summary

### Latest Test Run

```
✅ All Tests Passed: 4/4
✅ Health Check: OK
✅ Classify: 2 emails processed, each with 1 label
✅ Summarize: Summary generated
✅ Draft: Reply generated
```

**Service URL:** `https://cloud-inference-service-29216080826.us-central1.run.app`

**Status:** 🟢 Live and Working

---

## 📚 Documentation

- **`TEST_SCRIPTS_GUIDE.md`** - Chi tiết đầy đủ về tất cả test scripts
- **`QUICK_START.md`** - Hướng dẫn deploy và test
- **`README.md`** - Project documentation

---

## 🔧 Custom Usage

### Test với Custom URL

```powershell
$URL = "https://your-service-url.run.app"
.\test_all_endpoints.ps1 -ServiceUrl $URL
```

### Test Specific Endpoint với Custom Data

```powershell
# Custom classify test
$data = @{emails = @(@{id="test", content="Your email..."})} | ConvertTo-Json
$response = Invoke-WebRequest -Uri "$URL/classify" -Method Post -ContentType "application/json" -Body $data
$response.Content | ConvertFrom-Json

# Custom summarize test
$content = "Your email content..."
$response = Invoke-WebRequest -Uri "$URL/summarize" -Method Post -ContentType "text/plain" -Body $content
$response.Content | ConvertFrom-Json

# Custom draft test
$content = "Email to reply to..."
$response = Invoke-WebRequest -Uri "$URL/draft" -Method Post -ContentType "text/plain" -Body $content
$response.Content | ConvertFrom-Json
```

---

## ✅ Features

- ✅ **Batch Classification:** 1-100 emails per request
- ✅ **Single Label:** Mỗi email chỉ 1 label với score cao nhất
- ✅ **Email Summarization:** Tóm tắt nội dung email
- ✅ **Draft Reply:** Tự động tạo draft reply
- ✅ **Gzip Compression:** Tối ưu network transfer
- ✅ **JSON Format:** Request & Response đều JSON (classify) hoặc plain text (summarize/draft)

---

## 🎉 Success Criteria

Tất cả tests pass khi:

1. ✅ Health check returns `200 OK`
2. ✅ Classify returns exactly 1 label per email
3. ✅ Summarize returns non-empty summary
4. ✅ Draft returns non-empty draft reply
5. ✅ All responses in correct JSON format
6. ✅ All HTTP status codes are `200`

---

**Last Verified:** $(Get-Date -Format 'yyyy-MM-dd HH:mm')

**Status:** 🟢 All Systems Operational


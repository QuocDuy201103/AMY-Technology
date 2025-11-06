# Testing Guide

This guide covers different ways to test the Cloud-based Inference service.

## 1. Unit Tests

Run the unit tests locally:

```bash
go test ./...
```

Or with verbose output:

```bash
go test -v ./...
```

## 2. Local Testing

### Prerequisites

Set environment variables:

**PowerShell:**
```powershell
$env:DEEPSPEAK_API_KEY = "sk-bbbaa627589b4a338e2a3e010a3c11b5"
$env:DEEPSPEAK_API_URL = "https://api.deepseek.com"
$env:DEEPSEEK_MODEL = "deepseek-chat"  # Optional, defaults to deepseek-chat
$env:PORT = "8080"  # Optional, defaults to 8080
```

**Bash/Linux:**
```bash
export DEEPSPEAK_API_KEY="sk-bbbaa627589b4a338e2a3e010a3c11b5"
export DEEPSPEAK_API_URL="https://api.deepseek.com"
export DEEPSEEK_MODEL="deepseek-chat"  # Optional
export PORT="8080"  # Optional
```

### Start the Server

```bash
go run .
```

You should see:
```
2025/11/06 XX:XX:XX Using DEEPSPEAK_API_URL: https://api.deepseek.com
2025/11/06 XX:XX:XX DEEPSPEAK_API_KEY is configured (length: XX)
2025/11/06 XX:XX:XX Server starting on port 8080
```

### Test Endpoints

#### Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

#### Summarize Endpoint

```bash
curl -X POST http://localhost:8080/summarize \
  -H "Content-Type: text/html" \
  -d "<html><body>Hello from DeepSeek! This is a test email about a meeting scheduled for tomorrow at 3 PM.</body></html>"
```

Expected: Gzip-compressed JSON with summary
```json
{"summary":"Meeting scheduled for tomorrow at 3 PM"}
```

#### Classify Endpoint

```bash
curl -X POST http://localhost:8080/classify \
  -H "Content-Type: text/html" \
  -d "<html><body>URGENT: Please review the attached document immediately. This is a high-priority work item.</body></html>"
```

Expected response:
```json
{
  "labels": [
    {"label": "urgent", "score": 0.95},
    {"label": "work", "score": 0.88}
  ]
}
```

#### Draft Reply Endpoint

```bash
curl -X POST http://localhost:8080/draft \
  -H "Content-Type: text/html" \
  -d "<html><body>Hi, I wanted to follow up on our conversation from last week. Could you please send me the updated proposal?</body></html>"
```

Expected response:
```json
{
  "draft": "Thank you for your email. I will review the proposal and get back to you soon."
}
```

## 3. Testing Deployed Cloud Run Service

### Get Service URL

```bash
pulumi stack output serviceUrl
```

Or from gcloud:
```bash
gcloud run services describe cloud-inference-service-9c8569c \
  --project=cloud-based-inference \
  --region=us-central1 \
  --format="value(status.url)"
```

### Test Health Endpoint

curl -sS "$SERVICE_URL/health"

```bash
curl https://cloud-inference-service-9c8569c-w6w4ky3k7q-uc.a.run.app/health
# https://cloud-inference-service-9c8569c-w6w4ky3k7q-uc.a.run.app
```

### Test Summarize

```bash
curl -X POST https://cloud-inference-service-9c8569c-w6w4ky3k7q-uc.a.run.app/summarize \
  -H "Content-Type: text/html" \
  -d "<html><body>Test email content here</body></html>"
```

### Test Classify

```bash
curl -X POST https://YOUR-SERVICE-URL/classify \
  -H "Content-Type: text/html" \
  -d "<html><body>Urgent work email</body></html>"
```

### Test Draft

```bash
curl -X POST https://YOUR-SERVICE-URL/draft \
  -H "Content-Type: text/html" \
  -d "<html><body>Original email content</body></html>"
```

## 4. View Logs
gcloud run logs read cloud-inference-service --project=cloud-based-inference --region=us-central1


### Local Logs

Logs appear in the terminal where you ran `go run .`

### Cloud Run Logs

View recent logs:
```bash
gcloud run services logs read cloud-inference-service-9c8569c \
  --project=cloud-based-inference \
  --region=us-central1 \
  --limit=50
```

Stream logs in real-time:
```bash
gcloud run services logs tail cloud-inference-service-9c8569c \
  --project=cloud-based-inference \
  --region=us-central1
```

## 5. Testing with Different Models

You can test with different DeepSeek models by setting `DEEPSEEK_MODEL`:

```bash
export DEEPSEEK_MODEL="deepseek-coder"  # For code-related tasks
go run .
```

Or update in Pulumi:
```bash
pulumi config set deepseekModel deepseek-coder
pulumi up
```

## 6. Testing with Gemini

You can exercise the same endpoints via Gemini by setting Gemini envs locally (or wiring the server to pick Gemini based on env if desired):

```powershell
$env:GEMINI_API_KEY = "your-gemini-key"
$env:GEMINI_API_URL = "https://generativelanguage.googleapis.com/v1beta"  # optional
$env:GEMINI_MODEL = "models/gemini-1.5-flash"  # optional
```

Then call the same REST endpoints; the Gemini client returns the same response shapes.

Direct Gemini sanity test:
```bash
curl -sS "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=$GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents":[{"role":"user","parts":[{"text":"Say hello"}]}]
  }'
```

## 7. Troubleshooting

### Error: "DEEPSPEAK_API_KEY environment variable is required"
- Make sure you've set the environment variable before running `go run .`

### Error: "request failed: Post ... remote error: tls: unrecognized name"
- Check that `DEEPSPEAK_API_URL` is set to `https://api.deepseek.com` (not `api.deepspeak.com`)

### Error: "unexpected status code: 401"
- Your API key is invalid or expired. Get a new key from DeepSeek.

### Error: "unexpected status code: 429"
- Rate limit exceeded. Wait a moment and try again.

### Error: "model did not return valid JSON for classification"
- The model sometimes returns text instead of JSON. This is expected behavior - the classify endpoint may need refinement.

### 400 Bad Request
- Check that you're sending `Content-Type: text/html` header
- Ensure the request body contains valid HTML

## 8. Example Test Script

Create a file `test.sh`:

```bash
#!/bin/bash

BASE_URL="${1:-http://localhost:8080}"

echo "Testing Health..."
curl -s "$BASE_URL/health" | jq .

echo -e "\nTesting Summarize..."
curl -s -X POST "$BASE_URL/summarize" \
  -H "Content-Type: text/html" \
  -d "<html><body>Meeting at 3 PM tomorrow</body></html>" | gunzip | jq .

echo -e "\nTesting Classify..."
curl -s -X POST "$BASE_URL/classify" \
  -H "Content-Type: text/html" \
  -d "<html><body>URGENT: Review this document</body></html>" | jq .

echo -e "\nTesting Draft..."
curl -s -X POST "$BASE_URL/draft" \
  -H "Content-Type: text/html" \
  -d "<html><body>Hi, can you send the proposal?</body></html>" | jq .
```

Run it:
```bash
chmod +x test.sh
./test.sh                    # Test local
./test.sh https://YOUR-URL    # Test deployed
```


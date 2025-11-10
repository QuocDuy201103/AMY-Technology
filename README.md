# Cloud-based Inference Service

A Go-based email processing service deployed on Google Cloud Run, providing NLP capabilities through the Deepseek API.

## Features

- **POST /summarize** - Summarizes email content (returns gzip-compressed JSON)
- **POST /classify** - Classifies email content with labels
- **POST /draft** - Generates AI-powered draft replies

## Architecture

- **Language**: Go 1.21
- **Framework**: net/http with gorilla/mux
- **Infrastructure**: Pulumi (TypeScript) for GCP Cloud Run
- **API Client**: Deepseek API integration with retry logic

## Prerequisites

- Go 1.21+
- Node.js 18+ (for Pulumi)
- Docker
- Google Cloud SDK
- Pulumi CLI

## Setup

### 1. Install Dependencies

```bash
# Go dependencies
go mod download

# Pulumi dependencies
npm install
```

### 2. Configure Pulumi

```bash
# Login to Pulumi
pulumi login

# Set configuration
pulumi config set projectId your-gcp-project-id
pulumi config set region us-central1
pulumi config set deepseekApiUrl https://api.deepseek.com
pulumi config set --secret deepseekApiKey your-api-key
```

### 3. Build Docker Image

```bash
# Build the Docker image
docker build -t gcr.io/your-project-id/cloud-inference:latest .
# docker build -t gcr.io/cloud-based-inference/cloud-inference:latest .

# Push to Google Container Registry
docker push gcr.io/your-project-id/cloud-inference:latest
# docker push gcr.io/cloud-based-inference/cloud-inference:latest
```

### 4. Deploy Infrastructure

```bash
# Preview changes
pulumi preview

# Deploy
pulumi up
```

**For detailed GCP deployment instructions in Vietnamese, see [GCP_DEPLOYMENT.md](GCP_DEPLOYMENT.md)**

**For quick deployment guide for siftly-backend-dev project, see [DEPLOY_SIFTLY.md](DEPLOY_SIFTLY.md)**

## Local Development

### Run the Server

**PowerShell:**
```powershell
$env:DEEPSEEK_API_KEY = "sk-bbbaa627589b4a338e2a3e010a3c11b5"
$env:DEEPSEEK_API_URL = "https://api.deepseek.com"
$env:DEEPSEEK_MODEL = "deepseek-chat"  # Optional
go run .
```

**Bash/Linux:**
```bash
export DEEPSEEK_API_KEY="sk-your-api-key-here"
export DEEPSEEK_API_URL="https://api.deepseek.com"
export DEEPSEEK_MODEL="deepseek-chat"  # Optional
go run .
```

### Run Tests

```bash
go test ./...
```

### Quick Test Examples

```bash
# Health check
curl http://localhost:8080/health

# Summarize
curl -X POST http://localhost:8080/summarize \
  -H "Content-Type: text/html" \
  -d "<html><body>Your email content here</body></html>"

# Classify
curl -X POST http://localhost:8080/classify \
  -H "Content-Type: text/html" \
  -d "<html><body>Your email content here</body></html>"

# Draft
curl -X POST http://localhost:8080/draft \
  -H "Content-Type: text/html" \
  -d "<html><body>Your email content here</body></html>"
```

**For detailed testing instructions, see [TESTING.md](TESTING.md)**

## Environment Variables

 - `DEEPSEEK_API_KEY` (required) - API key for DeepSeek API
 - `DEEPSEEK_API_URL` (optional) - Base URL for DeepSeek API (default: https://api.deepseek.com)
 - `DEEPSEEK_MODEL` (optional) - Chat model name (default: deepseek-chat)
- `PORT` (optional) - Server port (default: 8080)
 - `GEMINI_API_KEY` (optional) - API key for Google Generative Language API
 - `GEMINI_API_URL` (optional) - Base URL for Gemini API (default: https://generativelanguage.googleapis.com/v1beta)
 - `GEMINI_MODEL` (optional) - Model path (default: models/gemini-1.5-flash)

## Project Structure

```
.
├── main.go                    # Main server application
├── deepseek_client.go        # Deepseek API client
├── deepseek_client_test.go   # Client unit tests
├── go.mod                     # Go module definition
├── Dockerfile                 # Docker build configuration
├── index.ts                   # Pulumi infrastructure code
├── Pulumi.yaml               # Pulumi project configuration
├── README.md                 # This file
├── GCP_DEPLOYMENT.md         # GCP deployment guide (Vietnamese)
└── TESTING.md                # Testing instructions
```

## API Client Features

The `DeepseekClient` includes:
- Automatic retries with exponential backoff (up to 3 retries)
- Timeout handling (30 seconds default)
- Error handling with structured API errors
- JSON response parsing

## Middleware

- **CORS** - Cross-Origin Resource Sharing support
- **Logging** - Request/response logging with timing
- **JSON Error Handling** - Consistent error response format
- **Panic Recovery** - Graceful error handling

## License

MIT


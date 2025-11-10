package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Server holds the application dependencies
type Server struct {
	client *DeepseekClient
}

// NewServer creates a new server instance
func NewServer() *Server {
	baseURL := os.Getenv("DEEPSEEK_API_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
		log.Printf("Using default DEEPSEEK_API_URL: %s", baseURL)
	} else {
		log.Printf("Using DEEPSEEK_API_URL: %s", baseURL)
	}

	apiKey := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY environment variable is required")
	}
	log.Printf("DEEPSEEK_API_KEY is configured (length: %d)", len(apiKey))

	return &Server{
		client: NewDeepseekClient(baseURL, apiKey),
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// JSONError writes an error response as JSON (with gzip compression)
func JSONError(w http.ResponseWriter, message string, statusCode int) {
	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	// Set status code first
	w.WriteHeader(statusCode)
	// Use gzip compression for error responses too
	if err := writeGzipJSON(w, errorResp); err != nil {
		// Fallback to uncompressed JSON if gzip fails
		w.Header().Set("Content-Type", "application/json")
		w.Header().Del("Content-Encoding") // Remove gzip header if set
		json.NewEncoder(w).Encode(errorResp)
	}
}

// CORS middleware
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, ww.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// JSONRecovery middleware for panic recovery
func JSONRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				JSONError(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// readRequestBody reads the request body, handling gzip decompression
func readRequestBody(r *http.Request) ([]byte, error) {
	var reader io.Reader = r.Body
	
	// Check if content is gzip compressed
	if r.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}
	
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// writeGzipJSON writes JSON response with gzip compression
func writeGzipJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")

	gz := gzip.NewWriter(w)
	defer gz.Close()

	return json.NewEncoder(gz).Encode(data)
}

// SummarizeHandler handles POST /summarize
func (s *Server) SummarizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := readRequestBody(r)
	if err != nil {
		JSONError(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	content := string(bodyBytes)
	if strings.TrimSpace(content) == "" {
		JSONError(w, "Email content is required", http.StatusBadRequest)
		return
	}

	summary, err := s.client.SummarizeEmail(content)
	if err != nil {
		log.Printf("Error calling Deepseek API for summarize: %v", err)
		// Log detailed error for debugging, but return generic message to client
		JSONError(w, "Failed to summarize email", http.StatusInternalServerError)
		return
	}

	if err := writeGzipJSON(w, summary); err != nil {
		log.Printf("Error writing response: %v", err)
		JSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// BatchClassifyRequest represents the batch classification request
type BatchClassifyRequest struct {
	Emails []EmailRequest `json:"emails"`
}

// ClassificationResult represents the classification result for a single email
type ClassificationResult struct {
	ID      string                 `json:"id"`
	Labels  []ClassificationLabel `json:"labels"`
}

// BatchClassifyResponse represents the batch classification response
type BatchClassifyResponse struct {
	Results []ClassificationResult `json:"results"`
}

// ClassifyHandler handles POST /classify
func (s *Server) ClassifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type must be application/json
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" && !strings.HasPrefix(contentType, "application/json;") {
		JSONError(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Read and decompress request body
	bodyBytes, err := readRequestBody(r)
	if err != nil {
		JSONError(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse JSON request
	var batchReq BatchClassifyRequest
	if err := json.Unmarshal(bodyBytes, &batchReq); err != nil {
		JSONError(w, fmt.Sprintf("Invalid JSON format: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(batchReq.Emails) == 0 {
		JSONError(w, "At least one email is required", http.StatusBadRequest)
		return
	}

	if len(batchReq.Emails) > 100 {
		JSONError(w, "Maximum 100 emails allowed per request", http.StatusBadRequest)
		return
	}

	// Validate each email
	for i, email := range batchReq.Emails {
		if strings.TrimSpace(email.ID) == "" {
			JSONError(w, fmt.Sprintf("Email ID is required for email at index %d", i), http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(email.Content) == "" {
			JSONError(w, fmt.Sprintf("Email content is required for email at index %d", i), http.StatusBadRequest)
			return
		}
	}

	// Process batch classification
	results, err := s.client.ClassifyEmailsBatch(batchReq.Emails)
	if err != nil {
		log.Printf("Error calling Deepseek API for batch classify: %v", err)
		JSONError(w, "Failed to classify emails", http.StatusInternalServerError)
		return
	}

	// Build response with only ID and classification result
	response := BatchClassifyResponse{
		Results: make([]ClassificationResult, len(results)),
	}
	for i, result := range results {
		response.Results[i] = ClassificationResult{
			ID:     result.ID,
			Labels: result.Labels,
		}
	}

	// Send compressed JSON response
	if err := writeGzipJSON(w, response); err != nil {
		log.Printf("Error writing response: %v", err)
		JSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DraftHandler handles POST /draft
func (s *Server) DraftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := readRequestBody(r)
	if err != nil {
		JSONError(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	content := string(bodyBytes)
	if strings.TrimSpace(content) == "" {
		JSONError(w, "Email content is required", http.StatusBadRequest)
		return
	}

	draft, err := s.client.DraftReply(content)
	if err != nil {
		log.Printf("Error calling Deepseek API for draft: %v", err)
		JSONError(w, "Failed to generate draft reply", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(draft); err != nil {
		log.Printf("Error writing response: %v", err)
		JSONError(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	server := NewServer()

	router := mux.NewRouter()

	// Apply middleware
	router.Use(JSONRecovery)
	router.Use(Logging)
	router.Use(CORS)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	// API endpoints
	router.HandleFunc("/summarize", server.SummarizeHandler).Methods("POST")
	router.HandleFunc("/classify", server.ClassifyHandler).Methods("POST")
	router.HandleFunc("/draft", server.DraftHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

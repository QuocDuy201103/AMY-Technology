package main

import (
    "compress/gzip"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gorilla/mux"
)

type Server struct {
    client *OpenAIClient
}

func NewServer() *Server {
    baseURL := os.Getenv("OPENAI_API_URL")
    if baseURL == "" {
        baseURL = "https://api.openai.com"
        log.Printf("Using default OPENAI_API_URL: %s", baseURL)
    } else {
        log.Printf("Using OPENAI_API_URL: %s", baseURL)
    }

    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("OPENAI_API_KEY environment variable is required")
    }
    log.Printf("OPENAI_API_KEY is configured (length: %d)", len(apiKey))

    return &Server{client: NewOpenAIClient(baseURL, apiKey)}
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

func JSONError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    _ = json.NewEncoder(w).Encode(ErrorResponse{Error: http.StatusText(statusCode), Message: message})
}

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

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(ww, r)
        duration := time.Since(start)
        log.Printf("%s %s %d %v", r.Method, r.URL.Path, ww.statusCode, duration)
    })
}

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

func readRequestBody(r *http.Request) (string, error) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return "", err
    }
    return string(body), nil
}

func writeGzipJSON(w http.ResponseWriter, data interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Encoding", "gzip")
    gz := gzip.NewWriter(w)
    defer gz.Close()
    return json.NewEncoder(gz).Encode(data)
}

func (s *Server) SummarizeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    content, err := readRequestBody(r)
    if err != nil {
        JSONError(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    if strings.TrimSpace(content) == "" {
        JSONError(w, "Email content is required", http.StatusBadRequest)
        return
    }
    summary, err := s.client.SummarizeEmail(content)
    if err != nil {
        log.Printf("Error calling OpenAI API for summarize: %v", err)
        JSONError(w, "Failed to summarize email", http.StatusInternalServerError)
        return
    }
    if err := writeGzipJSON(w, summary); err != nil {
        log.Printf("Error writing response: %v", err)
        JSONError(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}

func (s *Server) ClassifyHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    content, err := readRequestBody(r)
    if err != nil {
        JSONError(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    if strings.TrimSpace(content) == "" {
        JSONError(w, "Email content is required", http.StatusBadRequest)
        return
    }
    classification, err := s.client.ClassifyEmail(content)
    if err != nil {
        log.Printf("Error calling OpenAI API for classify: %v", err)
        JSONError(w, "Failed to classify email", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(classification); err != nil {
        log.Printf("Error writing response: %v", err)
        JSONError(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}

func (s *Server) DraftHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    content, err := readRequestBody(r)
    if err != nil {
        JSONError(w, "Failed to read request body", http.StatusBadRequest)
        return
    }
    if strings.TrimSpace(content) == "" {
        JSONError(w, "Email content is required", http.StatusBadRequest)
        return
    }
    draft, err := s.client.DraftReply(content)
    if err != nil {
        log.Printf("Error calling OpenAI API for draft: %v", err)
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
    router.Use(JSONRecovery)
    router.Use(Logging)
    router.Use(CORS)
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }).Methods("GET")
    router.HandleFunc("/summarize", server.SummarizeHandler).Methods("POST")
    router.HandleFunc("/classify", server.ClassifyHandler).Methods("POST")
    router.HandleFunc("/draft", server.DraftHandler).Methods("POST")
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("OpenAI server starting on port %s", port)
    if err := http.ListenAndServe(":"+port, router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}



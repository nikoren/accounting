package main

import (
	"accounting/internal/auth"
	"accounting/internal/config"
	"accounting/internal/httpapi"
	"accounting/internal/infrastructure/db/migrations"
	"accounting/internal/infrastructure/db/uow"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"accounting/internal/domain/ports"
	"accounting/internal/services"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/time/rate"
)

const (
	maxRequestSize  = 1 << 20 // 1MB
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 10 * time.Second
	// Rate limiting
	requestsPerSecond = 100
	burstSize         = 200
)

// JWTVerifierAdapter adapts auth.JWTMinter to httpapi.TokenVerifier
type JWTVerifierAdapter struct {
	minter *auth.JWTMinter
}

func (a *JWTVerifierAdapter) VerifyToken(token string) (any, error) {
	return a.minter.VerifyToken(token)
}

// metrics tracks server metrics
type metrics struct {
	mu            sync.RWMutex
	requestsTotal int64
	errorsTotal   int64
	lastError     error
	startTime     time.Time
	// New metrics
	requestDuration   atomic.Int64
	responseSize      atomic.Int64
	activeConnections atomic.Int32
	rateLimitHits     atomic.Int64
}

func (m *metrics) incrementRequests() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsTotal++
}

func (m *metrics) incrementErrors(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorsTotal++
	m.lastError = err
}

func (m *metrics) getStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return map[string]interface{}{
		"uptime_seconds":     time.Since(m.startTime).Seconds(),
		"requests_total":     m.requestsTotal,
		"errors_total":       m.errorsTotal,
		"last_error":         m.lastError,
		"avg_duration_ms":    float64(m.requestDuration.Load()) / float64(m.requestsTotal),
		"total_response_mb":  float64(m.responseSize.Load()) / (1024 * 1024),
		"active_connections": m.activeConnections.Load(),
		"rate_limit_hits":    m.rateLimitHits.Load(),
	}
}

// gzipResponseWriter wraps http.ResponseWriter to track response size
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	status int
	size   int64
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	size, err := w.Writer.Write(b)
	w.size += int64(size)
	return size, err
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	//TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
	// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize SQLite database
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Apply migrations
	if err := migrations.ApplyMigrations(db); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	// Create unit of work factory
	uowFactory := func() (ports.UnitOfWork, error) {
		uow := uow.NewUnitOfWorkSQL(db)
		if err := uow.Begin(); err != nil {
			return nil, err
		}
		return uow, nil
	}

	// Create render service
	renderSvc := services.NewRenderService()

	// Create split service
	splitSvc := services.NewSplitService(uowFactory, renderSvc)

	// Create JWT minter with users from config
	configUsers := cfg.GetUsersMap()
	users := make(map[string]auth.User, len(configUsers))
	for k, v := range configUsers {
		users[k] = auth.User{Username: v.Username, Password: v.Password}
	}
	jwtMinter, err := auth.NewJWTMinter(users)
	if err != nil {
		log.Fatalf("Failed to create JWT minter: %v", err)
	}

	// Create token verifier adapter
	tokenVerifier := &JWTVerifierAdapter{minter: jwtMinter}

	// Create split handler
	splitHandler := httpapi.NewSplitHandler(splitSvc, tokenVerifier)

	// Initialize metrics
	metrics := &metrics{
		startTime: time.Now(),
	}

	// Create rate limiter
	limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burstSize)

	// Create router
	mux := http.NewServeMux()

	// Register auth routes
	jwtMinter.Mount(mux)

	// Register split routes
	mux.HandleFunc("GET /splits/{id}", splitHandler.LoadSplitHandler)
	mux.HandleFunc("POST /splits/{id}/finalize", splitHandler.FinalizeSplitHandler)
	mux.HandleFunc("POST /documents", splitHandler.CreateDocumentHandler)
	mux.HandleFunc("PATCH /documents/{id}", splitHandler.UpdateDocumentMetadataHandler)
	mux.HandleFunc("DELETE /documents/{id}", splitHandler.DeleteDocumentHandler)
	mux.HandleFunc("GET /documents/{id}/download", splitHandler.DownloadDocumentHandler)
	mux.HandleFunc("POST /pages/move", splitHandler.MovePagesHandler)

	// Register metrics endpoint
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.getStats())
	})

	// Create middleware chain
	handler := chain(
		recoveryMiddleware,
		loggingMiddleware,
		requestIDMiddleware,
		metricsMiddleware(metrics),
		rateLimitMiddleware(limiter, metrics),
		compressionMiddleware,
	)(mux)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :%d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// rateLimitMiddleware implements rate limiting
func rateLimitMiddleware(limiter *rate.Limiter, metrics *metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				metrics.rateLimitHits.Add(1)
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// compressionMiddleware adds gzip compression
func compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldCompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
		}

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzw, r)
	})
}

// shouldCompress determines if the response should be compressed
func shouldCompress(r *http.Request) bool {
	// Skip compression for small responses
	if r.ContentLength > 0 && r.ContentLength < 1024 {
		return false
	}

	// Check if client accepts gzip
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return acceptEncoding != "" && acceptEncoding != "identity"
}

// metricsMiddleware tracks request metrics
func metricsMiddleware(m *metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			m.incrementRequests()
			m.activeConnections.Add(1)
			defer m.activeConnections.Add(-1)

			// Create a response writer that tracks size
			ww := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         w,
			}

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Milliseconds()
			m.requestDuration.Add(duration)
			m.responseSize.Add(ww.size)

			if ww.status >= 400 {
				m.incrementErrors(fmt.Errorf("request failed with status %d", ww.status))
			}
		})
	}
}

// chain creates a middleware chain
func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// loggingMiddleware logs information about each request using structured logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := newResponseWriter(w)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		log.Printf("request completed, method: %s, path: %s, status: %d, duration: %v, request_id: %s",
			r.Method, r.URL.Path, ww.status, duration, r.Context().Value("request_id"),
		)
	})
}

// recoveryMiddleware recovers from panics and returns 500
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered, error: %v, request_id: %s", err, r.Context().Value("request_id"))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requestIDMiddleware adds a unique request ID to each request
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = time.Now().Format("20060102150405.000000000")
		}
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// md5websrv.go - webservice to upload file and show MD5 Sum of it.
//
//	go build -ldflags '-w -s' -o MultiChecksumWeb .
//	./MultiChecksumWeb
//	go to http://localhost:8080/
package main

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

// define a FileObject
type file struct {
	Name        string // name of file
	ContentType string // content-type
	Size        int64  // size in bytes
	MD5         []byte // [16]byte md5 sum
	SHA1        []byte // [20]byte sha1 sum
	SHA224      []byte // [28]byte sha224 sum
	SHA256      []byte // [32]byte sha256 sum
	SHA512      []byte // [64]byte sha521 sum
}

// constants and variables:
const (
	maxUploadSize   = 100 << 20 // 100 MB
	shutdownTimeout = 5 * time.Second
	rateLimit       = 10 // requests per minute per IP
	rateBurst       = 10 // max burst size
)

// Rate limiter per IP
var (
	limiters  sync.Map // map[string]*rate.Limiter
	cleanupAt time.Time
)

// Custom template functions
var templateDir = os.Getenv("TEMPLATE_DIR")

func init() {
	if templateDir == "" {
		templateDir = "/tmpl/"
	}
}

var funcMap = template.FuncMap{
	"divf": func(a, b float64) float64 {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"float64": func(v int64) float64 {
		return float64(v)
	},
}

var templates = template.Must(template.New("").Funcs(funcMap).ParseGlob(templateDir + "*"))

// getLimiter returns a rate limiter for the given IP address
func getLimiter(ip string) *rate.Limiter {
	// Clean up old limiters periodically
	if time.Now().After(cleanupAt) {
		cleanupAt = time.Now().Add(10 * time.Minute)
		limiters.Range(func(key, value interface{}) bool {
			limiters.Delete(key)
			return true
		})
	}

	limiter, exists := limiters.Load(ip)
	if !exists {
		// Create new limiter: 10 requests per minute, burst of 10
		l := rate.NewLimiter(rate.Limit(rateLimit/60.0), rateBurst)
		limiters.Store(ip, l)
		return l
	}
	return limiter.(*rate.Limiter)
}

// rateLimitMiddleware applies rate limiting per IP
func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", 60.0/rateLimit))
			http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// shows the upload form
func upHandler(w http.ResponseWriter, r *http.Request) {
	// Set security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000")

	// Validate path and method: only GET / is allowed
	if r.URL.Path != "/" || r.Method != http.MethodGet {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if err := templates.ExecuteTemplate(w, "upload.html", nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("template execution failed", "handler", "upHandler", "error", err)
	}
}

// takes upload request and processes it
func doHandler(w http.ResponseWriter, r *http.Request) {
	// Set security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// get multipart-file from request
	mpf, mpHeader, err := r.FormFile("file")
	if err != nil {
		// Differenziertes Error Handling
		if errors.Is(err, http.ErrMissingFile) {
			http.Error(w, "No file provided", http.StatusBadRequest)
		} else if errors.Is(err, http.ErrNotMultipart) {
			http.Error(w, "Request must be multipart/form-data", http.StatusBadRequest)
		} else if errors.Is(err, http.ErrBodyTooLarge) {
			http.Error(w, "File too large (max 100MB)", http.StatusPayloadTooLarge)
		} else {
			http.Error(w, "Error retrieving file", http.StatusInternalServerError)
		}
		return
	}
	defer mpf.Close()

	// create new Checksum handles
	md5 := md5.New()
	sha1 := sha1.New()
	sha224 := sha256.New224()
	sha256 := sha256.New()
	sha512 := sha512.New()

	// create a MultiWriter to write to all handles at once
	mw := io.MultiWriter(md5, sha1, sha224, sha256, sha512)

	// stream the file directly to hash writers
	size, err := io.Copy(mw, mpf)
	if err != nil {
		http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
		slog.Error("file read failed", "filename", mpHeader.Filename, "size", size, "error", err)
		return
	}

	// Log successful file upload
	slog.Info("file uploaded", "name", mpHeader.Filename, "size", size, "content_type", mpHeader.Header.Get("Content-Type"))

	// get checksums
	md5sum := md5.Sum(nil)
	shasum := sha1.Sum(nil)
	sha224sum := sha224.Sum(nil)
	sha256sum := sha256.Sum(nil)
	sha512sum := sha512.Sum(nil)

	// parse uploaded data into my FileObject
	myFileObj := file{
		Name:        html.EscapeString(mpHeader.Filename), // XSS protection
		ContentType: html.EscapeString(mpHeader.Header.Get("Content-Type")),
		Size:        size,
		MD5:         md5sum,
		SHA1:        shasum,
		SHA224:      sha224sum,
		SHA256:      sha256sum,
		SHA512:      sha512sum,
	}

	// Execute template
	if err := templates.ExecuteTemplate(w, "download.html", myFileObj); err != nil {
		slog.Error("template execution failed", "handler", "doHandler", "error", err)
		// Check if headers were already written
		if !w.Header().Written() {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	hostPort := fmt.Sprintf("0.0.0.0:%s", port)

	http.HandleFunc("/", upHandler)
	http.HandleFunc("/up/", upHandler)
	http.HandleFunc("/do/", doHandler)
	http.HandleFunc("/health", healthHandler)

	server := &http.Server{
		Addr:    hostPort,
		Handler: http.DefaultServeMux,
	}

	// Signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		log.Printf("Shutting down server gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting MultiChecksumWeb on %s", hostPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("ListenAndServe: ", err)
	}
	log.Printf("Server stopped")
	// Initialize cleanup time
	cleanupAt = time.Now().Add(10 * time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rateLimitMiddleware(upHandler))
	mux.HandleFunc("/up/", rateLimitMiddleware(upHandler))
	mux.HandleFunc("/do/", rateLimitMiddleware(doHandler))
	mux.HandleFunc("/health", healthHandler) // health check without rate limiting

	slog.Info("starting server", "address", hostPort)
	server := &http.Server{
		Addr:         hostPort,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	slog.Error("server failed", "error", server.ListenAndServe())
	os.Exit(1)

	log.Printf("Starting MultiChecksumWeb on %s (rate limited to %d req/min per IP)", hostPort, rateLimit)
	log.Fatal(server.ListenAndServe())
}

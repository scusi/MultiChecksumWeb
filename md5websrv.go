// md5websrv.go - webservice to upload file and show MD5 Sum of it.
//
//  go build -ldflags '-w -s' -o MultiChecksumWeb .
//  ./MultiChecksumWeb
//  go to http://localhost:8080/
//
package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"html"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
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
	SHA512      []byte // [64]byte sha512 sum
}

// constants and variables:
const (
	maxUploadSize = 100 << 20 // 100 MB
)

// Custom template functions
var templateDir = os.Getenv("TEMPLATE_DIR")
if templateDir == "" {
	templateDir = "/tmpl/"
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

// shows the upload form
func upHandler(w http.ResponseWriter, r *http.Request) {
	// Set security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000")

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
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

	slog.Info("starting server", "address", hostPort)
	server := &http.Server{
		Addr:         hostPort,
		Handler:      nil,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	slog.Error("server failed", "error", server.ListenAndServe())
	os.Exit(1)
}

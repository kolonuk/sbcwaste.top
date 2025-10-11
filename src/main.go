package main

import (
	"compress/gzip"
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// _ "net/http/pprof"

	"github.com/chromedp/chromedp"
)

// Global variable to hold the allocator context
var allocatorContext context.Context

// findChromium returns the path to the google-chrome executable, or an empty string if not found.
func findChromium() string {
	for _, path := range []string{
		"google-chrome",
		"chromium-browser",
		"chromium",
	} {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}
	// Return empty string if no browser is found
	return ""
}

func main() {
	browserPath := findChromium()
	if browserPath == "" {
		log.Fatalf("No Chrome or Chromium browser found. Please install one.")
	}

	// Create a new chromedp allocator
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.Flag("no-sandbox", true), // Running as root requires this
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36`),
	)
	var cancel context.CancelFunc
	allocatorContext, cancel = chromedp.NewExecAllocator(context.Background(), opts...)

	// Set up a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run the graceful shutdown routine in a separate goroutine
	go func() {
		<-stop
		log.Println("Shutting down gracefully...")
		cancel()
		os.Exit(0)
	}()

	// Create a new ServeMux to handle routing.
	mux := http.NewServeMux()

	// Register handlers for each route.
	mux.Handle("/", http.HandlerFunc(serveIndex))
	mux.Handle("/static/", Gzip(cacheControlMiddleware(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))))
	mux.Handle("/.well-known/security.txt", http.HandlerFunc(serveSecurityTxt))
	mux.Handle("/search-address", http.HandlerFunc(SearchAddressHandler))
	mux.Handle("/health", http.HandlerFunc(healthCheckHandler))
	mux.Handle("/api/costs", http.HandlerFunc(BillingHandler))
	mux.Handle("/api/waste", http.HandlerFunc(WasteCollection))

	// The default handler for waste collection lookups. This will catch any requests
	// that don't match the other handlers, which is the desired behavior for
	// the /<uprn>/<format> endpoint.
	mux.Handle("/", http.HandlerFunc(WasteCollection))

	// Chain the middleware. The request will pass through the rate limiter first,
	// then the security headers handler, and finally to the router.
	handler := rateLimit(securityHeadersMiddleware(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler, // Use the chained handler
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("http.ListenAndServe: %v\n", err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	// Only set cache headers for static content in non-development environments.
	if os.Getenv("APP_ENV") != "development" {
		// Cache for 1 hour.
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	http.ServeFile(w, r, "static/index.html")
}

func serveSecurityTxt(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/.well-known/security.txt")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status": "ok"}`)); err != nil {
		log.Printf("Failed to write health check response: %v", err)
	}
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next.ServeHTTP(gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("APP_ENV") != "development" {
			// Cache for 1 day.
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add various security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		// Only add the HSTS header in non-development environments
		if os.Getenv("APP_ENV") != "development" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

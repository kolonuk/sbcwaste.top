package main

import (
	"compress/gzip"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
	// _ "net/http/pprof"
)

func main() {
	// Set up a channel to listen for OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run the graceful shutdown routine in a separate goroutine
	go func() {
		<-stop
		log.Println("Shutting down gracefully...")
		shutdownChromedp()
		os.Exit(0)
	}()

	// Create a new ServeMux to handle routing.
	mux := http.NewServeMux()

	// Register handlers for each route.
	mux.Handle("/.well-known/security.txt", http.HandlerFunc(serveSecurityTxt))
	mux.Handle("/search-address", http.HandlerFunc(SearchAddressHandler))
	mux.Handle("/health", http.HandlerFunc(healthCheckHandler))
	// Add the new file server handler.
	fileServer := Gzip(cacheControlMiddleware(http.FileServer(http.Dir("./static"))))
	mux.Handle("/", securityHeadersMiddleware(rootHandler(fileServer)))

	// Chain the middleware. The request will pass through the rate limiter first,
	// then the security headers handler, and finally to the router.
	handler := rateLimit(mux)

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

	log.Printf("Starting server on port %s", strings.ReplaceAll(port, "\n", "")) // #nosec G706 -- port is from env var, numeric-only after defaulting, newlines stripped
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("http.ListenAndServe: %v\n", err)
	}
}

func rootHandler(fileServer http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.Trim(r.URL.Path, "/")
		pathSegments := strings.Split(path, "/")

		// If it's the root or doesn't look like a UPRN path, serve static files
		if r.URL.Path == "/" || (len(pathSegments) > 0 && !regexp.MustCompile(`^[0-9]{1,20}$`).MatchString(pathSegments[0])) {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Otherwise, handle as a waste collection request
		WasteCollection(w, r)
	}
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

package main

import (
	"compress/gzip"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "net/http/pprof"

	"github.com/chromedp/chromedp"
)

// Global variable to hold the allocator context
var allocatorContext context.Context

func main() {
	// Create a new chromedp allocator
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath("/usr/bin/chromium"),
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

	http.HandleFunc("/", gzipMiddleware(router))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("http.ListenAndServe: %v\n", err)
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	// The GKE Ingress controller automatically adds a /healthz endpoint,
	// which returns 200 if the app is running. However, we want to be able
	// to run this locally, and also to have a more meaningful health check,
	// so we'll create our own /health endpoint.
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "static/index.html")
		return
	}

	// Serve static files
	if strings.HasPrefix(r.URL.Path, "/static/") {
		fs := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
		fs.ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/search-address" {
		SearchAddressHandler(w, r)
		return
	}
	if r.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	// For any other path, assume it's a waste collection request.
	// This will handle /<uprn>/json and /<uprn>/ics
	WasteCollection(w, r)
}


func gzipMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			w.Header().Set("Content-Encoding", "gzip")
			next(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
		} else {
			next(w, r)
		}
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	*gzip.Writer
}

// Write calls the Write method on the gzip.Writer, which is what you want for the gzip middleware.
func (w gzipResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

// Header calls the Header method on the embedded http.ResponseWriter.
func (w gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader calls the WriteHeader method on the embedded http.ResponseWriter.
func (w gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

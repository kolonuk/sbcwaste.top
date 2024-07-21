package main

import (
	"compress/gzip"
	"log"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"
)

func main() {
	http.HandleFunc("/", gzipMiddleware(WasteCollection))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("http.ListenAndServe: %v\n", err)
	}
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

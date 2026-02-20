package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// visitor stores a rate limiter for each visitor and the last time they were seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// visitors stores a map of visitors by IP address.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// init runs a background goroutine to clean up old entries from the visitors map.
func init() {
	go cleanupVisitors()
}

// getVisitor returns the rate limiter for the current visitor.
func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// Allow 10 requests per minute, with a burst of 5.
		limiter := rate.NewLimiter(rate.Every(time.Minute/10), 5)
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors periodically removes old entries from the visitors map.
func cleanupVisitors() {
	for {
		// Wait for 1 minute before next cleanup.
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range visitors {
			// If a visitor hasn't been seen for 3 minutes, remove them.
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// rateLimit is a middleware that limits requests per IP address.
func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do not rate limit the main page, the help page, or static assets
		if r.URL.Path == "/" || r.URL.Path == "/api/help" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// Get the IP address for the request.
		// The `X-Forwarded-For` header is the standard for identifying the
		// originating IP address of a client connecting through a proxy like
		// the one used by Google Cloud Run.
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			// If the header is not present, fall back to RemoteAddr.
			// This is useful for local development.
			var err error
			ip, _, err = net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				log.Printf("could not parse RemoteAddr: %v", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			// X-Forwarded-For can be a comma-separated list of IPs. The first one is the client's.
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		}

		// Get the rate limiter for the current IP address.
		limiter := getVisitor(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		// If the request is allowed, pass it to the next handler.
		next.ServeHTTP(w, r)
	})
}
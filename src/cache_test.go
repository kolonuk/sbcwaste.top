package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSqliteCache(t *testing.T) {
	// Create a new cache
	cache, err := NewSqliteCache()
	if err != nil {
		t.Fatalf("Failed to create sqlite cache: %v", err)
	}
	defer cache.Close()
	defer os.Remove("./sbcwaste.db")

	// Test Set and Get
	/*
		type Collection struct {
			Type            string   `json:"type"`
			CollectionDates []string `json:"CollectionDates"`
		}
		type Collections struct {
			Collections []Collection `json:"collections"`
			Address     string       `json:"address"`
		}
	*/
	// Using a simple JSON string to test the byte storage directly
	testValue := []byte(`{"address": "Test Address", "collections": [{"type": "Refuse", "CollectionDates": ["2025-01-01"]}]}`)

	err = cache.Set("12345", testValue, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	cachedBytes, created, err := cache.Get("12345")
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if cachedBytes == nil {
		t.Fatal("Cache returned nil")
	}
	if created.IsZero() {
		t.Fatal("Expected created timestamp, but it was zero")
	}
	if string(cachedBytes) != string(testValue) {
		t.Errorf("Expected value '%s', got '%s'", string(testValue), string(cachedBytes))
	}

	// Test cache expiration
	err = cache.Set("54321", testValue, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	cachedBytes, _, err = cache.Get("54321")
	if err != nil && err.Error() != "sql: no rows in result set" {
		t.Fatalf("Failed to get expired cache: %v", err)
	}
	if cachedBytes != nil {
		t.Fatal("Expected expired cache to be nil")
	}
}

func TestNewCache(t *testing.T) {
	// Test development environment
	os.Setenv("APP_ENV", "development")
	cache, err := NewCache(context.Background())
	if err != nil {
		t.Fatalf("Failed to create cache for development: %v", err)
	}
	if _, ok := cache.(*SqliteCache); !ok {
		t.Fatal("Expected SqliteCache for development")
	}
	defer os.Remove("./sbcwaste.db")

	// Test production environment (Firestore)
	os.Setenv("APP_ENV", "production")
	os.Setenv("PROJECT_ID", "dummy-project-for-testing") // Firestore client needs a project ID
	// This will attempt to create a Firestore client. We expect an error in a local
	// test environment where ADC (Application Default Credentials) are not available.
	_, err = NewCache(context.Background())
	if err == nil {
		t.Fatal("Expected an error when creating Firestore cache without proper credentials, but got nil")
	}
	// A more specific check could be to ensure the error is related to credentials,
	// but for now, just checking for a non-nil error is sufficient.
	os.Unsetenv("PROJECT_ID") // Clean up env var

	// Test self-hosted environment (Redis): REDIS_ADDR set should take
	// priority over Firestore regardless of APP_ENV, as long as it isn't
	// "development".
	addr := startFakeRedisServer(t)
	os.Setenv("REDIS_ADDR", addr)
	defer os.Unsetenv("REDIS_ADDR")

	cache, err = NewCache(context.Background())
	if err != nil {
		t.Fatalf("Failed to create cache for self-hosted deployment: %v", err)
	}
	defer cache.Close()
	if _, ok := cache.(*RedisCache); !ok {
		t.Fatal("Expected RedisCache when REDIS_ADDR is set")
	}
}

// startFakeRedisServer runs a minimal in-process RESP server backed by a
// plain map, just enough to exercise RedisCache's GET/SET encoding without
// needing a real Redis instance in CI. It ignores TTL (SET's EX argument) —
// that's Redis's own responsibility, not RedisCache's, so it isn't retested
// here.
func startFakeRedisServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start fake redis server: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	store := map[string]string{}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				r := bufio.NewReader(conn)
				for {
					args, err := readRESPCommand(r)
					if err != nil {
						return
					}
					if len(args) == 0 {
						continue
					}
					switch strings.ToUpper(args[0]) {
					case "AUTH":
						conn.Write([]byte("+OK\r\n"))
					case "SET":
						if len(args) < 3 {
							conn.Write([]byte("-ERR wrong number of arguments\r\n"))
							continue
						}
						store[args[1]] = args[2]
						conn.Write([]byte("+OK\r\n"))
					case "GET":
						if len(args) < 2 {
							conn.Write([]byte("-ERR wrong number of arguments\r\n"))
							continue
						}
						v, ok := store[args[1]]
						if !ok {
							conn.Write([]byte("$-1\r\n"))
							continue
						}
						conn.Write([]byte("$" + strconv.Itoa(len(v)) + "\r\n" + v + "\r\n"))
					default:
						conn.Write([]byte("-ERR unknown command\r\n"))
					}
				}
			}()
		}
	}()

	return ln.Addr().String()
}

// readRESPCommand reads one RESP array-of-bulk-strings command, the only
// shape RedisCache.do ever sends.
func readRESPCommand(r *bufio.Reader) ([]string, error) {
	header, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	header = strings.TrimRight(header, "\r\n")
	if len(header) == 0 || header[0] != '*' {
		return nil, errors.New("not a RESP array command")
	}
	n, err := strconv.Atoi(header[1:])
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		lenLine, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		lenLine = strings.TrimRight(lenLine, "\r\n")
		argLen, err := strconv.Atoi(strings.TrimPrefix(lenLine, "$"))
		if err != nil {
			return nil, err
		}
		data := make([]byte, argLen+2)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		args = append(args, string(data[:argLen]))
	}
	return args, nil
}

func TestRedisCache(t *testing.T) {
	addr := startFakeRedisServer(t)
	os.Setenv("REDIS_ADDR", addr)
	defer os.Unsetenv("REDIS_ADDR")

	cache, err := NewRedisCache()
	if err != nil {
		t.Fatalf("Failed to create redis cache: %v", err)
	}
	defer cache.Close()

	testValue := []byte(`{"address": "Test Address"}`)
	if err := cache.Set("12345", testValue, time.Hour); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	cachedBytes, created, err := cache.Get("12345")
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if string(cachedBytes) != string(testValue) {
		t.Errorf("Expected value '%s', got '%s'", string(testValue), string(cachedBytes))
	}
	if created.IsZero() {
		t.Fatal("Expected created timestamp, but it was zero")
	}

	// Cache miss
	missBytes, missCreated, err := cache.Get("does-not-exist")
	if err != nil {
		t.Fatalf("Expected no error on cache miss, got: %v", err)
	}
	if missBytes != nil {
		t.Fatal("Expected nil bytes for cache miss")
	}
	if !missCreated.IsZero() {
		t.Fatal("Expected zero time for cache miss")
	}
}

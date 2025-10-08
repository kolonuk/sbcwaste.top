package main

import (
	"context"
	"os"
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
	collections := &Collections{
		Collections: []Collection{
			{Type: "Refuse", CollectionDates: []string{"2025-01-01"}},
		},
		Address: "Test Address",
	}
	err = cache.Set("12345", collections, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	cachedCollections, err := cache.Get("12345")
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if cachedCollections == nil {
		t.Fatal("Cache returned nil")
	}
	if cachedCollections.Address != "Test Address" {
		t.Errorf("Expected address 'Test Address', got '%s'", cachedCollections.Address)
	}

	// Test cache expiration
	err = cache.Set("54321", collections, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}
	cachedCollections, err = cache.Get("54321")
	if err != nil && err.Error() != "sql: no rows in result set" {
		t.Fatalf("Failed to get expired cache: %v", err)
	}
	if cachedCollections != nil {
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
}
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	_ "github.com/mattn/go-sqlite3"
)

// Cache interface defines the methods for a cache
// It now stores raw bytes to be more generic.
type Cache interface {
	Get(key string) ([]byte, time.Time, error)
	Set(key string, value []byte, expiration time.Duration) error
	Close() error
}

// hashKey creates a SHA256 hash of the key.
// This is important for Firestore which cannot accept "/" in document IDs (e.g. URLs).
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// SqliteCache is a cache implementation using SQLite
type SqliteCache struct {
	db *sql.DB
}

// NewSqliteCache creates a new SqliteCache
func NewSqliteCache() (*SqliteCache, error) {
	db, err := sql.Open("sqlite3", "./sbcwaste.db")
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	// Create table if it doesn't exist. This will create the table with the old schema if it's the first run.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cache (
			key TEXT PRIMARY KEY,
			value TEXT,
			expiration INTEGER
		)
	`)
	if err != nil {
		return nil, err
	}

	// Attempt to add the 'created' column to support migrations from the old schema.
	// Ignore the error if the column already exists.
	_, err = db.Exec("ALTER TABLE cache ADD COLUMN created INTEGER")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return nil, err
	}

	return &SqliteCache{db: db}, nil
}

// Get retrieves a value from the SQLite cache
func (c *SqliteCache) Get(key string) ([]byte, time.Time, error) {
	hashedKey := hashKey(key)
	// We might have legacy unhashed keys in the DB if this is an existing file,
	// but for simplicity/robustness going forward we only look up by hash.
	// If backward compatibility with non-hashed keys was strictly required we'd check both,
	// but user approved the plan implying new keys.

	row := c.db.QueryRow("SELECT value, expiration, created FROM cache WHERE key = ?", hashedKey)

	var value string
	var expiration int64
	var created int64
	err := row.Scan(&value, &expiration, &created)
	if err != nil {
		return nil, time.Time{}, err // This will be sql.ErrNoRows for a cache miss
	}

	if time.Now().Unix() > expiration {
		// Cache expired
		if _, err := c.db.Exec("DELETE FROM cache WHERE key = ?", hashedKey); err != nil {
			log.Printf("Failed to delete expired cache entry for key %s (hash: %s): %v", key, hashedKey, err)
		}
		return nil, time.Time{}, nil
	}

	return []byte(value), time.Unix(created, 0), nil
}

// Set adds a value to the SQLite cache
func (c *SqliteCache) Set(key string, value []byte, expiration time.Duration) error {
	hashedKey := hashKey(key)
	now := time.Now()
	expirationTime := now.Add(expiration).Unix()
	createdTime := now.Unix()

	_, err := c.db.Exec(
		"INSERT OR REPLACE INTO cache (key, value, expiration, created) VALUES (?, ?, ?, ?)",
		hashedKey, string(value), expirationTime, createdTime,
	)
	return err
}

// Close closes the SQLite database connection
func (c *SqliteCache) Close() error {
	return c.db.Close()
}

// FirestoreCache is a cache implementation using Firestore
type FirestoreCache struct {
	client     *firestore.Client
	collection *firestore.CollectionRef
}

// FirestoreCacheItem represents the structure of the data stored in Firestore
type FirestoreCacheItem struct {
	Value      string    `firestore:"value"`
	Expiration time.Time `firestore:"expiration"`
	Created    time.Time `firestore:"created"`
}

// NewFirestoreCache creates a new FirestoreCache
func NewFirestoreCache(ctx context.Context) (*FirestoreCache, error) {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Println("PROJECT_ID environment variable not set.")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	appEnv := os.Getenv("APP_ENV")
	collectionName := "sbcwaste_cache" // Default to production
	if appEnv == "test" {
		collectionName = "sbcwaste_cache_test"
	}

	collection := client.Collection(collectionName)
	log.Printf("Using Firestore collection: %s", collectionName)

	return &FirestoreCache{client: client, collection: collection}, nil
}

// Get retrieves a value from the Firestore cache
func (c *FirestoreCache) Get(key string) ([]byte, time.Time, error) {
	hashedKey := hashKey(key)
	doc, err := c.collection.Doc(hashedKey).Get(context.Background())
	if err != nil {
		// This will handle the case where the document doesn't exist (cache miss)
		return nil, time.Time{}, nil
	}

	var item FirestoreCacheItem
	if err := doc.DataTo(&item); err != nil {
		return nil, time.Time{}, err
	}

	if time.Now().After(item.Expiration) {
		// Cache expired, delete the document
		_, _ = c.collection.Doc(hashedKey).Delete(context.Background())
		return nil, time.Time{}, nil
	}

	// Assuming the value stored is just the string representation of bytes (e.g. JSON or whatever)
	return []byte(item.Value), item.Created, nil
}

// Set adds a value to the Firestore cache
func (c *FirestoreCache) Set(key string, value []byte, expiration time.Duration) error {
	hashedKey := hashKey(key)
	now := time.Now()
	item := FirestoreCacheItem{
		Value:      string(value),
		Expiration: now.Add(expiration),
		Created:    now,
	}

	_, err := c.collection.Doc(hashedKey).Set(context.Background(), item)
	return err
}

// Close closes the Firestore client connection
func (c *FirestoreCache) Close() error {
	return c.client.Close()
}

// NewCache is a factory function that returns the appropriate cache implementation
// based on the environment
func NewCache(ctx context.Context) (Cache, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "development" {
		log.Println("Using SQLite cache for local development")
		return NewSqliteCache()
	}

	log.Println("Using Firestore cache for cloud environment")
	return NewFirestoreCache(ctx)
}

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	_ "github.com/mattn/go-sqlite3"
)

// Cache interface defines the methods for a cache
type Cache interface {
	Get(key string) (*Collections, error)
	Set(key string, collections *Collections, expiration time.Duration) error
	Close() error
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

	return &SqliteCache{db: db}, nil
}

// Get retrieves a value from the SQLite cache
func (c *SqliteCache) Get(key string) (*Collections, error) {
	row := c.db.QueryRow("SELECT value, expiration FROM cache WHERE key = ?", key)

	var value string
	var expiration int64
	err := row.Scan(&value, &expiration)
	if err != nil {
		return nil, err // This will be sql.ErrNoRows for a cache miss
	}

	if time.Now().Unix() > expiration {
		// Cache expired
		c.db.Exec("DELETE FROM cache WHERE key = ?", key)
		return nil, nil
	}

	var collections Collections
	err = json.Unmarshal([]byte(value), &collections)
	if err != nil {
		return nil, err
	}

	return &collections, nil
}

// Set adds a value to the SQLite cache
func (c *SqliteCache) Set(key string, collections *Collections, expiration time.Duration) error {
	value, err := json.Marshal(collections)
	if err != nil {
		return err
	}

	expirationTime := time.Now().Add(expiration).Unix()

	_, err = c.db.Exec(
		"INSERT OR REPLACE INTO cache (key, value, expiration) VALUES (?, ?, ?)",
		key, string(value), expirationTime,
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
}

// NewFirestoreCache creates a new FirestoreCache
func NewFirestoreCache(ctx context.Context) (*FirestoreCache, error) {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		// In a real application, you might want to return an error here
		// or have a more robust way of getting the project ID.
		log.Println("PROJECT_ID environment variable not set.")
	}

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	collection := client.Collection("sbcwaste_cache")

	return &FirestoreCache{client: client, collection: collection}, nil
}

// Get retrieves a value from the Firestore cache
func (c *FirestoreCache) Get(key string) (*Collections, error) {
	doc, err := c.collection.Doc(key).Get(context.Background())
	if err != nil {
		// This will handle the case where the document doesn't exist (cache miss)
		return nil, nil
	}

	var item FirestoreCacheItem
	if err := doc.DataTo(&item); err != nil {
		return nil, err
	}

	if time.Now().After(item.Expiration) {
		// Cache expired, delete the document
		_, _ = c.collection.Doc(key).Delete(context.Background())
		return nil, nil
	}

	var collections Collections
	err = json.Unmarshal([]byte(item.Value), &collections)
	if err != nil {
		return nil, err
	}

	return &collections, nil
}

// Set adds a value to the Firestore cache
func (c *FirestoreCache) Set(key string, collections *Collections, expiration time.Duration) error {
	val, err := json.Marshal(collections)
	if err != nil {
		return err
	}

	item := FirestoreCacheItem{
		Value:      string(val),
		Expiration: time.Now().Add(expiration),
	}

	_, err = c.collection.Doc(key).Set(context.Background(), item)
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
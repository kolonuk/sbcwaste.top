package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
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
		return nil, err
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

// MemcachedCache is a cache implementation using Memcached
type MemcachedCache struct {
	client *memcache.Client
}

// NewMemcachedCache creates a new MemcachedCache
func NewMemcachedCache() (*MemcachedCache, error) {
	memcachedEndpoint := os.Getenv("MEMCACHED_DISCOVERY_ENDPOINT")
	if memcachedEndpoint == "" {
		memcachedEndpoint = "localhost:11211"
	}

	client := memcache.New(memcachedEndpoint)
	err := client.Ping()
	if err != nil {
		return nil, err
	}

	return &MemcachedCache{client: client}, nil
}

// Get retrieves a value from the Memcached cache
func (c *MemcachedCache) Get(key string) (*Collections, error) {
	item, err := c.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, err
	}

	var collections Collections
	err = json.Unmarshal(item.Value, &collections)
	if err != nil {
		return nil, err
	}

	return &collections, nil
}

// Set adds a value to the Memcached cache
func (c *MemcachedCache) Set(key string, collections *Collections, expiration time.Duration) error {
	val, err := json.Marshal(collections)
	if err != nil {
		return err
	}

	return c.client.Set(&memcache.Item{
		Key:        key,
		Value:      val,
		Expiration: int32(expiration.Seconds()),
	})
}

// Close is a no-op for the memcache client but is here to satisfy the interface
func (c *MemcachedCache) Close() error {
	return nil
}

// NewCache is a factory function that returns the appropriate cache implementation
// based on the environment
func NewCache(ctx context.Context) (Cache, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "development" {
		log.Println("Using SQLite cache for local development")
		return NewSqliteCache()
	}

	log.Println("Using Memcached cache for cloud environment")
	return NewMemcachedCache()
}
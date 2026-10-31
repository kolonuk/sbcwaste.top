package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2/google"
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
	projectID := strings.TrimSpace(os.Getenv("PROJECT_ID"))
	if projectID == "" {
		return nil, fmt.Errorf("PROJECT_ID environment variable not set")
	}
	if strings.HasPrefix(projectID, "dummy-") {
		return nil, fmt.Errorf("project ID %q is not a real Firestore project", projectID)
	}

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		_, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return nil, fmt.Errorf("firestore credentials not configured: %w", err)
		}
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

// RedisCache is a cache implementation backed by a Redis (or Redis-compatible,
// e.g. Valkey) server, talked to directly over RESP — no client library
// dependency, which keeps `go.mod` free of anything that would need cgo or a
// new go.sum entry. Intended for the self-hosted deployment (dadock), which
// has no GCP project and shouldn't need one just to cache waste-collection
// lookups for a low-traffic site.
type RedisCache struct {
	mu   sync.Mutex
	conn net.Conn
	r    *bufio.Reader
}

// redisEnvelope mirrors the (value, created) pair the Sqlite/Firestore
// caches store natively, since a plain Redis string value has nowhere else
// to carry the "created" timestamp alongside the cached bytes. Expiry itself
// is left to Redis's own key TTL (SET ... EX) rather than tracked here.
type redisEnvelope struct {
	Value   string `json:"value"`
	Created int64  `json:"created"`
}

// NewRedisCache creates a new RedisCache, connecting to REDIS_ADDR
// (host:port) and authenticating with REDIS_PASSWORD if set.
func NewRedisCache() (*RedisCache, error) {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		return nil, fmt.Errorf("REDIS_ADDR environment variable not set")
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second) // #nosec G704 -- REDIS_ADDR is an operator-set env var, not user input
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}

	c := &RedisCache{conn: conn, r: bufio.NewReader(conn)}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		if _, isNil, err := c.do("AUTH", password); err != nil || isNil {
			_ = conn.Close()
			return nil, fmt.Errorf("redis AUTH failed: %w", err)
		}
	}

	return c, nil
}

// do sends a single RESP command and returns its reply. isNil distinguishes
// a nil bulk reply ($-1, e.g. GET on a missing key) from a valid empty
// string, which a plain string return value can't.
func (c *RedisCache) do(args ...string) (value string, isNil bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var cmd strings.Builder
	fmt.Fprintf(&cmd, "*%d\r\n", len(args))
	for _, a := range args {
		fmt.Fprintf(&cmd, "$%d\r\n%s\r\n", len(a), a)
	}
	if _, err := io.WriteString(c.conn, cmd.String()); err != nil {
		return "", false, err
	}

	line, err := c.r.ReadString('\n')
	if err != nil {
		return "", false, err
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return "", false, fmt.Errorf("empty redis reply")
	}

	switch line[0] {
	case '+', ':': // simple string / integer
		return line[1:], false, nil
	case '-': // error
		return "", false, fmt.Errorf("redis error: %s", line[1:])
	case '$': // bulk string
		n, convErr := strconv.Atoi(line[1:])
		if convErr != nil {
			return "", false, fmt.Errorf("invalid redis bulk length %q: %w", line[1:], convErr)
		}
		if n == -1 {
			return "", true, nil
		}
		data := make([]byte, n+2) // payload + trailing CRLF
		if _, err := io.ReadFull(c.r, data); err != nil {
			return "", false, err
		}
		return string(data[:n]), false, nil
	default:
		return "", false, fmt.Errorf("unexpected redis reply type %q", line)
	}
}

// Get retrieves a value from the Redis cache. A missing/expired key is a
// cache miss, reported the same way as FirestoreCache: nil bytes, zero
// time, nil error.
func (c *RedisCache) Get(key string) ([]byte, time.Time, error) {
	hashedKey := hashKey(key)
	val, isNil, err := c.do("GET", hashedKey)
	if err != nil {
		return nil, time.Time{}, err
	}
	if isNil {
		return nil, time.Time{}, nil
	}

	var env redisEnvelope
	if err := json.Unmarshal([]byte(val), &env); err != nil {
		return nil, time.Time{}, err
	}
	return []byte(env.Value), time.Unix(env.Created, 0), nil
}

// Set adds a value to the Redis cache, relying on Redis's own TTL (SET ...
// EX) for expiry rather than checking it ourselves on read.
func (c *RedisCache) Set(key string, value []byte, expiration time.Duration) error {
	hashedKey := hashKey(key)
	env := redisEnvelope{Value: string(value), Created: time.Now().Unix()}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}

	seconds := int64(expiration.Seconds())
	if seconds < 1 {
		seconds = 1 // Redis EX requires a positive integer.
	}

	_, _, err = c.do("SET", hashedKey, string(data), "EX", strconv.FormatInt(seconds, 10))
	return err
}

// Close closes the Redis connection.
func (c *RedisCache) Close() error {
	return c.conn.Close()
}

// NewCache is a factory function that returns the appropriate cache implementation
// based on the environment:
//   - APP_ENV=development -> SQLite (local dev, no external services)
//   - REDIS_ADDR set       -> Redis (self-hosted/dadock deployment)
//   - otherwise            -> Firestore (Google Cloud deployment)
func NewCache(ctx context.Context) (Cache, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "development" {
		log.Println("Using SQLite cache for local development")
		return NewSqliteCache()
	}

	if addr := strings.TrimSpace(os.Getenv("REDIS_ADDR")); addr != "" {
		log.Println("Using Redis cache for self-hosted deployment")
		return NewRedisCache()
	}

	log.Println("Using Firestore cache for cloud environment")
	return NewFirestoreCache(ctx)
}

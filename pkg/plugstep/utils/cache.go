package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	_ "modernc.org/sqlite"
)

const (
	schemaVersion   = 1
	CacheTTL        = 15 * time.Minute
	CacheTTLShort   = 15 * time.Minute
	CacheTTLForever = 0
)

type Cache struct {
	db        *sql.DB
	namespace string
	mu        sync.RWMutex
}

var (
	globalDB   *sql.DB
	globalDBMu sync.Mutex
	caches     = map[string]*Cache{}
)

func initDB(serverDirectory string) (*sql.DB, error) {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()

	if globalDB != nil {
		return globalDB, nil
	}

	cacheDir := filepath.Join(serverDirectory, ".plugstep")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cacheDir, "cache.db")

	// Check schema version of existing db
	if _, err := os.Stat(dbPath); err == nil {
		db, err := sql.Open("sqlite", dbPath)
		if err == nil {
			var version int
			err := db.QueryRow("PRAGMA user_version").Scan(&version)
			db.Close()
			if err != nil || version != schemaVersion {
				log.Debug("Cache schema mismatch, rebuilding", "old", version, "new", schemaVersion)
				os.Remove(dbPath)
			}
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Set pragmas for performance
	_, err = db.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA cache_size=1000;
		PRAGMA temp_store=MEMORY;
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cache (
			namespace TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			timestamp INTEGER NOT NULL,
			ttl INTEGER NOT NULL,
			PRIMARY KEY (namespace, key)
		);
		CREATE INDEX IF NOT EXISTS idx_cache_namespace ON cache(namespace);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Set schema version
	_, err = db.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion))
	if err != nil {
		db.Close()
		return nil, err
	}

	globalDB = db
	return db, nil
}

// InitCacheDB initializes the cache database for a server directory
func InitCacheDB(serverDirectory string) error {
	_, err := initDB(serverDirectory)
	return err
}

// CloseCache closes the global database connection
func CloseCache() {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()
	if globalDB != nil {
		globalDB.Close()
		globalDB = nil
	}
	caches = map[string]*Cache{}
}

func InitCache(name string) *Cache {
	if cache, ok := caches[name]; ok {
		return cache
	}

	if globalDB == nil {
		log.Debug("Cache DB not initialized")
		return nil
	}

	cache := &Cache{
		db:        globalDB,
		namespace: name,
	}
	caches[name] = cache
	return cache
}

func GetCache(name string) *Cache {
	return caches[name]
}

func FlushCache(name string) error {
	if globalDB == nil {
		return nil
	}

	_, err := globalDB.Exec("DELETE FROM cache WHERE namespace = ?", name)
	if err != nil {
		return err
	}

	delete(caches, name)
	log.Info("Cache flushed", "namespace", name)
	return nil
}

func (c *Cache) Get(key string, dest interface{}) bool {
	if c == nil || c.db == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	var value string
	var timestamp, ttl int64
	err := c.db.QueryRow(
		"SELECT value, timestamp, ttl FROM cache WHERE namespace = ? AND key = ?",
		c.namespace, key,
	).Scan(&value, &timestamp, &ttl)

	if err != nil {
		return false
	}

	// Check TTL (0 = forever)
	if ttl > 0 {
		expiry := time.Unix(0, timestamp).Add(time.Duration(ttl))
		if time.Now().After(expiry) {
			// Expired, delete it
			go func() {
				c.mu.Lock()
				defer c.mu.Unlock()
				c.db.Exec("DELETE FROM cache WHERE namespace = ? AND key = ?", c.namespace, key)
			}()
			return false
		}
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return false
	}

	log.Debug("Cache hit", "namespace", c.namespace, "key", key)
	return true
}

// Set caches a value with the default short TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, CacheTTLShort)
}

// SetPermanent caches a value forever
func (c *Cache) SetPermanent(key string, value interface{}) {
	c.SetWithTTL(key, value, CacheTTLForever)
}

// SetWithTTL caches a value with a specific TTL (0 = forever)
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if c == nil || c.db == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	_, err = c.db.Exec(`
		INSERT OR REPLACE INTO cache (namespace, key, value, timestamp, ttl)
		VALUES (?, ?, ?, ?, ?)
	`, c.namespace, key, string(data), time.Now().UnixNano(), int64(ttl))

	if err != nil {
		log.Debug("Failed to write cache", "err", err)
	}
}

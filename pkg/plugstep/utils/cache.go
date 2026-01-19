package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
)

const (
	CacheTTL      = 15 * time.Minute
	CacheTTLShort = 15 * time.Minute
	CacheTTLForever = 0
)

type CacheEntry struct {
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	TTL       time.Duration   `json:"ttl"`
}

type Cache struct {
	dir string
}

var globalCaches = map[string]*Cache{}

func InitCache(name string) *Cache {
	if cache, ok := globalCaches[name]; ok {
		return cache
	}

	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Debug("Failed to get user cache directory", "err", err)
		return nil
	}

	cacheDir := filepath.Join(userCacheDir, "plugstep", name)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Debug("Failed to create cache directory", "err", err)
		return nil
	}

	cache := &Cache{dir: cacheDir}
	globalCaches[name] = cache
	return cache
}

func GetCache(name string) *Cache {
	return globalCaches[name]
}

func FlushCache(name string) error {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(userCacheDir, "plugstep", name)
	if err := os.RemoveAll(cacheDir); err != nil {
		return err
	}

	delete(globalCaches, name)
	log.Info("Cache flushed", "dir", cacheDir)
	return nil
}

func (c *Cache) cacheKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:8])
}

func (c *Cache) Get(key string, dest interface{}) bool {
	if c == nil {
		return false
	}

	path := filepath.Join(c.dir, c.cacheKey(key)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false
	}

	// TTL of 0 means forever, otherwise check expiry
	if entry.TTL > 0 && time.Since(entry.Timestamp) > entry.TTL {
		os.Remove(path)
		return false
	}

	if err := json.Unmarshal(entry.Data, dest); err != nil {
		return false
	}

	log.Debug("Cache hit", "key", key)
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
	if c == nil {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	entryData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	path := filepath.Join(c.dir, c.cacheKey(key)+".json")
	if err := os.WriteFile(path, entryData, 0644); err != nil {
		log.Debug("Failed to write cache", "err", err)
	}
}

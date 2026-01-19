package plugins

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
	CacheTTL = 15 * time.Minute
)

type CacheEntry struct {
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

type PluginCache struct {
	dir string
}

var globalCache *PluginCache

func InitCache() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Debug("Failed to get user cache directory", "err", err)
		return
	}

	cacheDir := filepath.Join(userCacheDir, "plugstep", "plugins")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Debug("Failed to create cache directory", "err", err)
		return
	}
	globalCache = &PluginCache{dir: cacheDir}
}

func GetCache() *PluginCache {
	return globalCache
}

func FlushCache() error {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(userCacheDir, "plugstep", "plugins")
	if err := os.RemoveAll(cacheDir); err != nil {
		return err
	}

	log.Info("Cache flushed", "dir", cacheDir)
	return nil
}

func (c *PluginCache) cacheKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:8])
}

func (c *PluginCache) Get(key string, dest interface{}) bool {
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

	if time.Since(entry.Timestamp) > CacheTTL {
		os.Remove(path)
		return false
	}

	if err := json.Unmarshal(entry.Data, dest); err != nil {
		return false
	}

	log.Debug("Cache hit", "key", key)
	return true
}

func (c *PluginCache) Set(key string, value interface{}) {
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

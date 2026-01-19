package utils

import (
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
)

const fileHashCacheName = "filehash"

type cachedHash struct {
	Hash  string `json:"hash"`
	Size  int64  `json:"size"`
	Mtime int64  `json:"mtime"`
}

func getFileHashCache() *Cache {
	return InitCache(fileHashCacheName)
}

func CalculateFileSHA256(filename string) (string, error) {
	return calculateFileHash(filename, "sha256", func(file *os.File) (string, error) {
		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", hasher.Sum(nil)), nil
	})
}

func CalculateFileSHA512(filename string) (string, error) {
	return calculateFileHash(filename, "sha512", func(file *os.File) (string, error) {
		hasher := sha512.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return "", err
		}
		return fmt.Sprintf("%x", hasher.Sum(nil)), nil
	})
}

func calculateFileHash(filename string, hashType string, hashFunc func(*os.File) (string, error)) (string, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return "", err
	}

	size := stat.Size()
	mtime := stat.ModTime().UnixNano()
	cacheKey := fmt.Sprintf("%s:%s:%d:%d", hashType, filename, size, mtime)

	cache := getFileHashCache()
	var cached cachedHash
	if cache != nil && cache.Get(cacheKey, &cached) {
		// Verify cached entry matches current file
		if cached.Size == size && cached.Mtime == mtime {
			return cached.Hash, nil
		}
	}

	// Calculate hash
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash, err := hashFunc(file)
	if err != nil {
		return "", err
	}

	// Cache the result permanently
	if cache != nil {
		cache.SetPermanent(cacheKey, cachedHash{
			Hash:  hash,
			Size:  size,
			Mtime: mtime,
		})
	}

	return hash, nil
}

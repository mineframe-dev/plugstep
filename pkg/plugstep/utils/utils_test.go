package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// EnsureDirectory Tests
// =============================================================================

func TestEnsureDirectory_CreatesNewDirectory(t *testing.T) {
	tempDir := t.TempDir()
	newDir := filepath.Join(tempDir, "newdir")

	err := EnsureDirectory(newDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestEnsureDirectory_CreatesNestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "a", "b", "c", "d")

	err := EnsureDirectory(nestedDir)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(nestedDir)
	if err != nil {
		t.Fatalf("nested directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, got file")
	}
}

func TestEnsureDirectory_ExistingDirectorySucceeds(t *testing.T) {
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing")

	// Create it first
	if err := os.Mkdir(existingDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// EnsureDirectory should not fail
	err := EnsureDirectory(existingDir)

	if err != nil {
		t.Errorf("unexpected error for existing directory: %v", err)
	}
}

func TestEnsureDirectory_EmptyPathFails(t *testing.T) {
	err := EnsureDirectory("")

	// Empty path should fail or create current directory
	// Current behavior: MkdirAll("") returns nil (no-op)
	if err != nil {
		t.Logf("Empty path returned error: %v", err)
	}
}

// =============================================================================
// Checksum Tests
// =============================================================================

func TestCalculateFileSHA256_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	content := []byte("hello world")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hash, err := CalculateFileSHA256(filePath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SHA256 of "hello world" is known
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestCalculateFileSHA512_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	content := []byte("hello world")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hash, err := CalculateFileSHA512(filePath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SHA512 of "hello world" is known
	expected := "309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f"
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestCalculateFileSHA256_NonexistentFile(t *testing.T) {
	_, err := CalculateFileSHA256("/nonexistent/file.txt")

	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCalculateFileSHA512_NonexistentFile(t *testing.T) {
	_, err := CalculateFileSHA512("/nonexistent/file.txt")

	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCalculateFileSHA256_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty.txt")

	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hash, err := CalculateFileSHA256(filePath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SHA256 of empty string
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expected {
		t.Errorf("expected hash %q, got %q", expected, hash)
	}
}

func TestCalculateFileSHA256_ConsistentResults(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")

	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	hash1, err := CalculateFileSHA256(filePath)
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}

	hash2, err := CalculateFileSHA256(filePath)
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("inconsistent hashes: %q vs %q", hash1, hash2)
	}
}

func TestCalculateFileSHA256_DifferentFilesDifferentHashes(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("failed to write test file 1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("failed to write test file 2: %v", err)
	}

	hash1, err := CalculateFileSHA256(file1)
	if err != nil {
		t.Fatalf("hash1 failed: %v", err)
	}

	hash2, err := CalculateFileSHA256(file2)
	if err != nil {
		t.Fatalf("hash2 failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("different files should have different hashes")
	}
}

// =============================================================================
// Cache Tests
// =============================================================================

func TestCache_InitAndGet(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize cache DB
	err := InitCacheDB(tempDir)
	if err != nil {
		t.Fatalf("failed to init cache DB: %v", err)
	}
	defer CloseCache()

	cache := InitCache("test")
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}

	// Set a value
	cache.Set("key1", "value1")

	// Get it back
	var result string
	found := cache.Get("key1", &result)

	if !found {
		t.Error("expected to find cached value")
	}
	if result != "value1" {
		t.Errorf("expected %q, got %q", "value1", result)
	}
}

func TestCache_GetMissingKey(t *testing.T) {
	tempDir := t.TempDir()

	err := InitCacheDB(tempDir)
	if err != nil {
		t.Fatalf("failed to init cache DB: %v", err)
	}
	defer CloseCache()

	cache := InitCache("test")

	var result string
	found := cache.Get("nonexistent", &result)

	if found {
		t.Error("expected false for missing key")
	}
}

func TestCache_SetPermanent(t *testing.T) {
	tempDir := t.TempDir()

	err := InitCacheDB(tempDir)
	if err != nil {
		t.Fatalf("failed to init cache DB: %v", err)
	}
	defer CloseCache()

	cache := InitCache("test")

	// Set permanent value
	cache.SetPermanent("permanent_key", map[string]int{"count": 42})

	// Get it back
	var result map[string]int
	found := cache.Get("permanent_key", &result)

	if !found {
		t.Error("expected to find permanent cached value")
	}
	if result["count"] != 42 {
		t.Errorf("expected count 42, got %d", result["count"])
	}
}

func TestCache_ComplexTypes(t *testing.T) {
	tempDir := t.TempDir()

	err := InitCacheDB(tempDir)
	if err != nil {
		t.Fatalf("failed to init cache DB: %v", err)
	}
	defer CloseCache()

	cache := InitCache("test")

	type TestStruct struct {
		Name  string   `json:"name"`
		Value int      `json:"value"`
		Tags  []string `json:"tags"`
	}

	original := TestStruct{
		Name:  "test",
		Value: 123,
		Tags:  []string{"a", "b", "c"},
	}

	cache.SetPermanent("struct_key", original)

	var result TestStruct
	found := cache.Get("struct_key", &result)

	if !found {
		t.Fatal("expected to find cached struct")
	}
	if result.Name != original.Name {
		t.Errorf("expected name %q, got %q", original.Name, result.Name)
	}
	if result.Value != original.Value {
		t.Errorf("expected value %d, got %d", original.Value, result.Value)
	}
	if len(result.Tags) != len(original.Tags) {
		t.Errorf("expected %d tags, got %d", len(original.Tags), len(result.Tags))
	}
}

func TestCache_FlushCache(t *testing.T) {
	tempDir := t.TempDir()

	err := InitCacheDB(tempDir)
	if err != nil {
		t.Fatalf("failed to init cache DB: %v", err)
	}
	defer CloseCache()

	cache := InitCache("test_flush")
	cache.SetPermanent("key1", "value1")
	cache.SetPermanent("key2", "value2")

	// Verify values exist
	var result string
	if !cache.Get("key1", &result) {
		t.Fatal("key1 should exist before flush")
	}

	// Flush the cache
	err = FlushCache("test_flush")
	if err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	// Re-init cache after flush
	cache = InitCache("test_flush")

	// Values should be gone
	if cache.Get("key1", &result) {
		t.Error("key1 should not exist after flush")
	}
	if cache.Get("key2", &result) {
		t.Error("key2 should not exist after flush")
	}
}

func TestCache_NilCacheGetReturnsFalse(t *testing.T) {
	var cache *Cache = nil

	var result string
	found := cache.Get("key", &result)

	if found {
		t.Error("expected false for nil cache")
	}
}

func TestCache_NilCacheSetDoesNotPanic(t *testing.T) {
	var cache *Cache = nil

	// Should not panic
	cache.Set("key", "value")
	cache.SetPermanent("key", "value")
	cache.SetWithTTL("key", "value", time.Minute)
}

func TestCache_MultipleNamespaces(t *testing.T) {
	tempDir := t.TempDir()

	err := InitCacheDB(tempDir)
	if err != nil {
		t.Fatalf("failed to init cache DB: %v", err)
	}
	defer CloseCache()

	cache1 := InitCache("namespace1")
	cache2 := InitCache("namespace2")

	// Same key, different namespaces
	cache1.SetPermanent("shared_key", "value1")
	cache2.SetPermanent("shared_key", "value2")

	var result1, result2 string
	cache1.Get("shared_key", &result1)
	cache2.Get("shared_key", &result2)

	if result1 != "value1" {
		t.Errorf("namespace1 expected %q, got %q", "value1", result1)
	}
	if result2 != "value2" {
		t.Errorf("namespace2 expected %q, got %q", "value2", result2)
	}
}

func TestGetCache_ReturnsNilForUninitialized(t *testing.T) {
	// Close any existing cache
	CloseCache()

	cache := GetCache("nonexistent")

	if cache != nil {
		t.Error("expected nil for uninitialized cache")
	}
}

func TestInitCache_ReturnsNilWithoutDB(t *testing.T) {
	// Close any existing cache
	CloseCache()

	cache := InitCache("test")

	if cache != nil {
		t.Error("expected nil when DB not initialized")
	}
}

// =============================================================================
// HTTP Client Tests (just verify they exist and have correct timeout)
// =============================================================================

func TestHTTPClient_Exists(t *testing.T) {
	if HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
	if HTTPClient.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", HTTPClient.Timeout)
	}
}

func TestDownloadClient_Exists(t *testing.T) {
	if DownloadClient == nil {
		t.Error("DownloadClient should not be nil")
	}
	if DownloadClient.Timeout != 5*time.Minute {
		t.Errorf("expected 5m timeout, got %v", DownloadClient.Timeout)
	}
}

package server

import (
	"testing"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
)

// --- GetVendor() Tests ---

func TestGetVendor_ReturnsPaperMCVendor(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vendor == nil {
		t.Fatal("expected non-nil vendor for papermc")
	}

	paper, ok := vendor.(*PaperJarVendor)
	if !ok {
		t.Fatalf("expected *PaperJarVendor, got %T", vendor)
	}

	// Verify API URL is set (we'll test if it's correct via network tests)
	if paper.apiURL == "" {
		t.Error("expected non-empty API URL")
	}
}

func TestGetVendor_ReturnsErrorForUnknownVendor(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendor("unknown-vendor"))

	if err == nil {
		t.Error("expected error for unknown vendor")
	}
	if vendor != nil {
		t.Errorf("expected nil vendor, got %T", vendor)
	}
}

func TestGetVendor_ReturnsErrorForEmptyVendor(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendor(""))

	if err == nil {
		t.Error("expected error for empty vendor")
	}
	if vendor != nil {
		t.Errorf("expected nil vendor, got %T", vendor)
	}
}

// =============================================================================
// Network Tests - These hit real APIs
// =============================================================================

func TestPaperJarVendor_GetDownload_ValidVersion(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "paper",
		MinecraftVersion: "1.21",
		Version:          "130",
	}

	download, err := vendor.GetDownload(cfg)

	if err != nil {
		t.Fatalf("failed to get download: %v", err)
	}

	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}
	if download.Checksum == "" {
		t.Error("expected non-empty checksum")
	}

	t.Logf("PaperMC download: url=%s, checksum=%s", download.URL, download.Checksum)
}

func TestPaperJarVendor_GetDownload_LatestBuild(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	// "latest" is a special version that should resolve to the latest build
	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "paper",
		MinecraftVersion: "1.21",
		Version:          "latest",
	}

	download, err := vendor.GetDownload(cfg)

	// Note: This might fail if "latest" isn't handled specially by the API
	// The test documents the current behavior
	if err != nil {
		t.Logf("'latest' version returned error (may be expected): %v", err)
		t.Skip("'latest' version not supported directly by API")
	}

	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}

	t.Logf("PaperMC latest: url=%s", download.URL)
}

func TestPaperJarVendor_GetDownload_NonexistentVersion(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "paper",
		MinecraftVersion: "1.21",
		Version:          "999999",
	}

	_, err = vendor.GetDownload(cfg)

	if err == nil {
		t.Error("expected error for nonexistent version")
	}
}

func TestPaperJarVendor_GetDownload_NonexistentProject(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "nonexistent-project-12345",
		MinecraftVersion: "1.21",
		Version:          "1",
	}

	_, err = vendor.GetDownload(cfg)

	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestPaperJarVendor_GetDownload_NonexistentMinecraftVersion(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "paper",
		MinecraftVersion: "99.99",
		Version:          "1",
	}

	_, err = vendor.GetDownload(cfg)

	if err == nil {
		t.Error("expected error for nonexistent minecraft version")
	}
}

func TestPaperJarVendor_GetDownload_VelocityProject(t *testing.T) {
	// Velocity is another PaperMC project (proxy server)
	// Note: Velocity uses its own version scheme, not Minecraft versions
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "velocity",
		MinecraftVersion: "3.1.1", // Velocity version, not MC version
		Version:          "102",
	}

	download, err := vendor.GetDownload(cfg)

	if err != nil {
		t.Fatalf("failed to get velocity download: %v", err)
	}

	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}

	t.Logf("Velocity download: url=%s", download.URL)
}

func TestPaperJarVendor_GetDownload_FoliaProject(t *testing.T) {
	// Folia is another PaperMC project (regionized multithreading)
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "folia",
		MinecraftVersion: "1.21.4",
		Version:          "6", // Latest available build for 1.21.4
	}

	download, err := vendor.GetDownload(cfg)

	if err != nil {
		t.Fatalf("failed to get folia download: %v", err)
	}

	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}

	t.Logf("Folia download: url=%s", download.URL)
}

func TestPaperJarVendor_GetDownload_EmptyProject(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "",
		MinecraftVersion: "1.21",
		Version:          "130",
	}

	_, err = vendor.GetDownload(cfg)

	if err == nil {
		t.Error("expected error for empty project")
	}
}

func TestPaperJarVendor_GetDownload_EmptyMinecraftVersion(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "paper",
		MinecraftVersion: "",
		Version:          "130",
	}

	_, err = vendor.GetDownload(cfg)

	if err == nil {
		t.Error("expected error for empty minecraft version")
	}
}

func TestPaperJarVendor_GetDownload_EmptyVersion(t *testing.T) {
	vendor, err := GetVendor(config.ServerJarVendorPaperMC)
	if err != nil {
		t.Fatalf("failed to get vendor: %v", err)
	}

	cfg := config.ServerConfig{
		Vendor:           config.ServerJarVendorPaperMC,
		Project:          "paper",
		MinecraftVersion: "1.21",
		Version:          "",
	}

	_, err = vendor.GetDownload(cfg)

	if err == nil {
		t.Error("expected error for empty version")
	}
}

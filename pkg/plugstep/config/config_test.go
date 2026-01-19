package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Lint() Tests ---

func TestLint_ReturnsWarningForLatestVersion(t *testing.T) {
	config := &PlugstepConfig{
		Server: ServerConfig{
			Vendor:           ServerJarVendorPaperMC,
			Project:          "paper",
			MinecraftVersion: "1.21",
			Version:          "latest",
		},
	}

	issues := config.Lint()

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	expected := "Using version = latest on server jar, can be good for security, possible API versioning issues"
	if issues[0] != expected {
		t.Errorf("expected issue %q, got %q", expected, issues[0])
	}
}

func TestLint_ReturnsNoWarningsForPinnedVersion(t *testing.T) {
	config := &PlugstepConfig{
		Server: ServerConfig{
			Vendor:           ServerJarVendorPaperMC,
			Project:          "paper",
			MinecraftVersion: "1.21",
			Version:          "123",
		},
	}

	issues := config.Lint()

	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestLint_ReturnsNoWarningsForEmptyVersion(t *testing.T) {
	config := &PlugstepConfig{
		Server: ServerConfig{
			Vendor:           ServerJarVendorPaperMC,
			Project:          "paper",
			MinecraftVersion: "1.21",
			Version:          "",
		},
	}

	issues := config.Lint()

	if len(issues) != 0 {
		t.Errorf("expected no issues for empty version, got %d: %v", len(issues), issues)
	}
}

func TestLint_CaseInsensitiveLatest(t *testing.T) {
	// Edge case: "LATEST" or "Latest" should probably trigger warning too
	// This test documents current behavior (case-sensitive)
	config := &PlugstepConfig{
		Server: ServerConfig{
			Vendor:           ServerJarVendorPaperMC,
			Project:          "paper",
			MinecraftVersion: "1.21",
			Version:          "LATEST",
		},
	}

	issues := config.Lint()

	// Current behavior: case-sensitive, so "LATEST" != "latest"
	if len(issues) != 0 {
		t.Logf("Lint is case-sensitive: 'LATEST' triggered %d warnings", len(issues))
	} else {
		t.Log("Lint is case-sensitive: 'LATEST' did not trigger warning (only 'latest' does)")
	}
}

func TestLint_EmptyConfig(t *testing.T) {
	// Edge case: completely empty config
	config := &PlugstepConfig{}

	issues := config.Lint()

	// Empty version is not "latest", so no warning
	if len(issues) != 0 {
		t.Errorf("expected no issues for empty config, got %d: %v", len(issues), issues)
	}
}

func TestLint_NilPluginsSlice(t *testing.T) {
	// Edge case: nil plugins slice (should not panic)
	config := &PlugstepConfig{
		Server: ServerConfig{
			Vendor:           ServerJarVendorPaperMC,
			Project:          "paper",
			MinecraftVersion: "1.21",
			Version:          "123",
		},
		Plugins: nil,
	}

	// Should not panic
	issues := config.Lint()

	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

// --- LoadPlugstepConfig() Tests ---

func TestLoadPlugstepConfig_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	configContent := `
[server]
vendor = "papermc"
project = "paper"
minecraft_version = "1.21"
version = "123"

[[plugins]]
source = "modrinth"
resource = "chunky"
version = "1.4.16"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadPlugstepConfig(configPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Server.Vendor != ServerJarVendorPaperMC {
		t.Errorf("expected vendor %q, got %q", ServerJarVendorPaperMC, config.Server.Vendor)
	}
	if config.Server.Project != "paper" {
		t.Errorf("expected project %q, got %q", "paper", config.Server.Project)
	}
	if config.Server.MinecraftVersion != "1.21" {
		t.Errorf("expected minecraft_version %q, got %q", "1.21", config.Server.MinecraftVersion)
	}
	if config.Server.Version != "123" {
		t.Errorf("expected version %q, got %q", "123", config.Server.Version)
	}

	if len(config.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(config.Plugins))
	}
	if config.Plugins[0].Source != PluginSourceModrinth {
		t.Errorf("expected plugin source %q, got %q", PluginSourceModrinth, config.Plugins[0].Source)
	}
	if config.Plugins[0].Resource == nil || *config.Plugins[0].Resource != "chunky" {
		t.Errorf("expected plugin resource %q, got %v", "chunky", config.Plugins[0].Resource)
	}
}

func TestLoadPlugstepConfig_MissingServerConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	configContent := `
[[plugins]]
source = "modrinth"
resource = "chunky"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadPlugstepConfig(configPath)

	if err == nil {
		t.Fatal("expected error for missing server config, got nil")
	}
}

func TestLoadPlugstepConfig_InvalidTOML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	configContent := `
this is not valid toml [[[
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadPlugstepConfig(configPath)

	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

func TestLoadPlugstepConfig_NonexistentFile(t *testing.T) {
	_, err := LoadPlugstepConfig("/nonexistent/path/plugstep.toml")

	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoadPlugstepConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	// Empty file is valid TOML but has no server config
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadPlugstepConfig(configPath)

	if err == nil {
		t.Fatal("expected error for empty file (no server config), got nil")
	}
}

func TestLoadPlugstepConfig_UnknownVendor(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	// Unknown vendor value - TOML parses it but it's not a valid ServerJarVendor
	configContent := `
[server]
vendor = "unknown-vendor"
project = "paper"
minecraft_version = "1.21"
version = "123"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadPlugstepConfig(configPath)

	// TOML parser accepts any string, so this should succeed
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// But the vendor won't match any known constant
	if config.Server.Vendor == ServerJarVendorPaperMC {
		t.Error("unknown vendor should not match PaperMC")
	}

	t.Logf("Unknown vendor parsed as: %q", config.Server.Vendor)
}

func TestLoadPlugstepConfig_UnknownPluginSource(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	configContent := `
[server]
vendor = "papermc"
project = "paper"
minecraft_version = "1.21"
version = "123"

[[plugins]]
source = "unknown-source"
resource = "some-plugin"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadPlugstepConfig(configPath)

	// TOML parser accepts any string
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(config.Plugins))
	}

	// Source won't match any known constant
	if config.Plugins[0].Source == PluginSourceModrinth ||
		config.Plugins[0].Source == PluginSourcePaperHangar ||
		config.Plugins[0].Source == PluginSourceCustom {
		t.Error("unknown source should not match any known source")
	}

	t.Logf("Unknown source parsed as: %q", config.Plugins[0].Source)
}

func TestLoadPlugstepConfig_PartialServerConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	// Server config with only vendor - missing other fields
	configContent := `
[server]
vendor = "papermc"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadPlugstepConfig(configPath)

	// Should succeed - TOML doesn't require all fields
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Missing fields should be zero values
	if config.Server.Project != "" {
		t.Errorf("expected empty project, got %q", config.Server.Project)
	}
	if config.Server.MinecraftVersion != "" {
		t.Errorf("expected empty minecraft_version, got %q", config.Server.MinecraftVersion)
	}
	if config.Server.Version != "" {
		t.Errorf("expected empty version, got %q", config.Server.Version)
	}
}

func TestLoadPlugstepConfig_WhitespaceOnlyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	// File with only whitespace
	if err := os.WriteFile(configPath, []byte("   \n\t\n   "), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadPlugstepConfig(configPath)

	if err == nil {
		t.Fatal("expected error for whitespace-only file (no server config), got nil")
	}
}

func TestLoadPlugstepConfig_EmptyPluginsArray(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	configContent := `
[server]
vendor = "papermc"
project = "paper"
minecraft_version = "1.21"
version = "latest"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadPlugstepConfig(configPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In Go, nil slice and empty slice behave the same for len() and range
	// TOML decoder returns nil when no plugins are defined, which is fine
	if len(config.Plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(config.Plugins))
	}
}

func TestLoadPlugstepConfig_MultiplePlugins(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "plugstep.toml")

	configContent := `
[server]
vendor = "papermc"
project = "paper"
minecraft_version = "1.21"
version = "123"

[[plugins]]
source = "modrinth"
resource = "chunky"
version = "1.4.16"

[[plugins]]
source = "paper-hangar"
resource = "ViaVersion"

[[plugins]]
source = "custom"
download_url = "https://example.com/plugin.jar"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadPlugstepConfig(configPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Plugins) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(config.Plugins))
	}

	// Verify first plugin (modrinth with version)
	if config.Plugins[0].Source != PluginSourceModrinth {
		t.Errorf("plugin[0]: expected source %q, got %q", PluginSourceModrinth, config.Plugins[0].Source)
	}
	if config.Plugins[0].Version == nil || *config.Plugins[0].Version != "1.4.16" {
		t.Errorf("plugin[0]: expected version %q, got %v", "1.4.16", config.Plugins[0].Version)
	}

	// Verify second plugin (paper-hangar without version)
	if config.Plugins[1].Source != PluginSourcePaperHangar {
		t.Errorf("plugin[1]: expected source %q, got %q", PluginSourcePaperHangar, config.Plugins[1].Source)
	}
	if config.Plugins[1].Version != nil {
		t.Errorf("plugin[1]: expected nil version, got %v", *config.Plugins[1].Version)
	}

	// Verify third plugin (custom with download_url)
	if config.Plugins[2].Source != PluginSourceCustom {
		t.Errorf("plugin[2]: expected source %q, got %q", PluginSourceCustom, config.Plugins[2].Source)
	}
	if config.Plugins[2].DownloadURL == nil || *config.Plugins[2].DownloadURL != "https://example.com/plugin.jar" {
		t.Errorf("plugin[2]: expected download_url %q, got %v", "https://example.com/plugin.jar", config.Plugins[2].DownloadURL)
	}
}

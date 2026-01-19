package commands

import (
	"testing"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
)

// =============================================================================
// parsePluginSpec Tests
// =============================================================================

func TestParsePluginSpec_ValidModrinthSpec(t *testing.T) {
	spec, err := parsePluginSpec("modrinth:luckperms")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Source != "modrinth" {
		t.Errorf("expected source %q, got %q", "modrinth", spec.Source)
	}
	if spec.Name != "luckperms" {
		t.Errorf("expected name %q, got %q", "luckperms", spec.Name)
	}
	if spec.Version != "" {
		t.Errorf("expected empty version, got %q", spec.Version)
	}
}

func TestParsePluginSpec_ValidHangarSpec(t *testing.T) {
	spec, err := parsePluginSpec("hangar:ViaVersion")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Source != "hangar" {
		t.Errorf("expected source %q, got %q", "hangar", spec.Source)
	}
	if spec.Name != "ViaVersion" {
		t.Errorf("expected name %q, got %q", "ViaVersion", spec.Name)
	}
}

func TestParsePluginSpec_ValidCustomSpec(t *testing.T) {
	spec, err := parsePluginSpec("custom:myplugin")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Source != "custom" {
		t.Errorf("expected source %q, got %q", "custom", spec.Source)
	}
	if spec.Name != "myplugin" {
		t.Errorf("expected name %q, got %q", "myplugin", spec.Name)
	}
}

func TestParsePluginSpec_WithVersion(t *testing.T) {
	spec, err := parsePluginSpec("modrinth:chunky@1.4.16")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Source != "modrinth" {
		t.Errorf("expected source %q, got %q", "modrinth", spec.Source)
	}
	if spec.Name != "chunky" {
		t.Errorf("expected name %q, got %q", "chunky", spec.Name)
	}
	if spec.Version != "1.4.16" {
		t.Errorf("expected version %q, got %q", "1.4.16", spec.Version)
	}
}

func TestParsePluginSpec_VersionWithAtSymbol(t *testing.T) {
	// Edge case: version contains @ symbol
	spec, err := parsePluginSpec("modrinth:plugin@1.0@beta")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only first @ is the delimiter
	if spec.Name != "plugin" {
		t.Errorf("expected name %q, got %q", "plugin", spec.Name)
	}
	if spec.Version != "1.0@beta" {
		t.Errorf("expected version %q, got %q", "1.0@beta", spec.Version)
	}
}

func TestParsePluginSpec_InvalidNoColon(t *testing.T) {
	_, err := parsePluginSpec("luckperms")

	if err == nil {
		t.Error("expected error for spec without colon")
	}
}

func TestParsePluginSpec_InvalidSource(t *testing.T) {
	_, err := parsePluginSpec("invalid:luckperms")

	if err == nil {
		t.Error("expected error for invalid source")
	}
}

func TestParsePluginSpec_EmptyName(t *testing.T) {
	_, err := parsePluginSpec("modrinth:")

	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestParsePluginSpec_EmptySpec(t *testing.T) {
	_, err := parsePluginSpec("")

	if err == nil {
		t.Error("expected error for empty spec")
	}
}

func TestParsePluginSpec_OnlyColon(t *testing.T) {
	_, err := parsePluginSpec(":")

	if err == nil {
		t.Error("expected error for only colon")
	}
}

func TestParsePluginSpec_NameWithSpecialChars(t *testing.T) {
	spec, err := parsePluginSpec("modrinth:my-plugin_v2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Name != "my-plugin_v2" {
		t.Errorf("expected name %q, got %q", "my-plugin_v2", spec.Name)
	}
}

// =============================================================================
// sourceToConfigSource Tests
// =============================================================================

func TestSourceToConfigSource_Modrinth(t *testing.T) {
	result := sourceToConfigSource("modrinth")

	if result != config.PluginSourceModrinth {
		t.Errorf("expected %q, got %q", config.PluginSourceModrinth, result)
	}
}

func TestSourceToConfigSource_Hangar(t *testing.T) {
	result := sourceToConfigSource("hangar")

	if result != config.PluginSourcePaperHangar {
		t.Errorf("expected %q, got %q", config.PluginSourcePaperHangar, result)
	}
}

func TestSourceToConfigSource_Custom(t *testing.T) {
	result := sourceToConfigSource("custom")

	if result != config.PluginSourceCustom {
		t.Errorf("expected %q, got %q", config.PluginSourceCustom, result)
	}
}

func TestSourceToConfigSource_Unknown(t *testing.T) {
	result := sourceToConfigSource("unknown")

	if result != "" {
		t.Errorf("expected empty string for unknown source, got %q", result)
	}
}

func TestSourceToConfigSource_Empty(t *testing.T) {
	result := sourceToConfigSource("")

	if result != "" {
		t.Errorf("expected empty string for empty source, got %q", result)
	}
}

// =============================================================================
// getSourceBadge Tests
// =============================================================================

func TestGetSourceBadge_Modrinth(t *testing.T) {
	badge := getSourceBadge("modrinth")

	if badge == "" {
		t.Error("expected non-empty badge for modrinth")
	}
	// Badge should contain MODRINTH text
	// Note: actual content includes ANSI styling codes
}

func TestGetSourceBadge_PaperHangar(t *testing.T) {
	badge := getSourceBadge("paper-hangar")

	if badge == "" {
		t.Error("expected non-empty badge for paper-hangar")
	}
}

func TestGetSourceBadge_Custom(t *testing.T) {
	badge := getSourceBadge("custom")

	if badge == "" {
		t.Error("expected non-empty badge for custom")
	}
}

func TestGetSourceBadge_Unknown(t *testing.T) {
	badge := getSourceBadge("unknown-source")

	if badge == "" {
		t.Error("expected non-empty badge for unknown source (should uppercase it)")
	}
}

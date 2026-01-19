package plugins

import (
	"testing"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
)

// --- GetSource() Tests ---

func TestGetSource_ReturnsModrinthSource(t *testing.T) {
	source := GetSource(config.PluginSourceModrinth)

	if source == nil {
		t.Fatal("expected non-nil source for modrinth")
	}

	modrinth, ok := source.(*ModrinthPluginSource)
	if !ok {
		t.Fatalf("expected *ModrinthPluginSource, got %T", source)
	}

	if modrinth.apiURL != "https://api.modrinth.com/v2" {
		t.Errorf("expected API URL %q, got %q", "https://api.modrinth.com/v2", modrinth.apiURL)
	}
}

func TestGetSource_ReturnsPaperHangarSource(t *testing.T) {
	source := GetSource(config.PluginSourcePaperHangar)

	if source == nil {
		t.Fatal("expected non-nil source for paper-hangar")
	}

	hangar, ok := source.(*PaperHangarPluginSource)
	if !ok {
		t.Fatalf("expected *PaperHangarPluginSource, got %T", source)
	}

	if hangar.apiURL != "https://hangar.papermc.io/api/v1" {
		t.Errorf("expected API URL %q, got %q", "https://hangar.papermc.io/api/v1", hangar.apiURL)
	}
}

func TestGetSource_ReturnsCustomSource(t *testing.T) {
	source := GetSource(config.PluginSourceCustom)

	if source == nil {
		t.Fatal("expected non-nil source for custom")
	}

	_, ok := source.(*CustomPluginSource)
	if !ok {
		t.Fatalf("expected *CustomPluginSource, got %T", source)
	}
}

func TestGetSource_ReturnsNilForUnknownSource(t *testing.T) {
	source := GetSource(config.PluginSource("unknown-source"))

	if source != nil {
		t.Errorf("expected nil for unknown source, got %T", source)
	}
}

// --- findModrinthVersion() Tests ---

func TestFindModrinthVersion_FindsMatchingVersion(t *testing.T) {
	versions := []ModrinthVersion{
		{VersionNumber: "1.0.0", Files: []ModrinthFile{}},
		{VersionNumber: "1.1.0", Files: []ModrinthFile{}},
		{VersionNumber: "1.2.0", Files: []ModrinthFile{}},
	}

	result := findModrinthVersion(versions, "1.1.0")

	if result == nil {
		t.Fatal("expected to find version 1.1.0")
	}
	if result.VersionNumber != "1.1.0" {
		t.Errorf("expected version %q, got %q", "1.1.0", result.VersionNumber)
	}
}

func TestFindModrinthVersion_ReturnsNilForMissingVersion(t *testing.T) {
	versions := []ModrinthVersion{
		{VersionNumber: "1.0.0", Files: []ModrinthFile{}},
		{VersionNumber: "1.1.0", Files: []ModrinthFile{}},
	}

	result := findModrinthVersion(versions, "2.0.0")

	if result != nil {
		t.Errorf("expected nil for missing version, got %v", result)
	}
}

func TestFindModrinthVersion_ReturnsNilForEmptyArray(t *testing.T) {
	versions := []ModrinthVersion{}

	result := findModrinthVersion(versions, "1.0.0")

	if result != nil {
		t.Errorf("expected nil for empty array, got %v", result)
	}
}

func TestFindModrinthVersion_ReturnsNilForNilSlice(t *testing.T) {
	var versions []ModrinthVersion = nil

	result := findModrinthVersion(versions, "1.0.0")

	if result != nil {
		t.Errorf("expected nil for nil slice, got %v", result)
	}
}

func TestFindModrinthVersion_IsCaseSensitive(t *testing.T) {
	versions := []ModrinthVersion{
		{VersionNumber: "v1.0.0", Files: []ModrinthFile{}},
	}

	result := findModrinthVersion(versions, "V1.0.0")

	if result != nil {
		t.Error("expected case-sensitive matching to return nil")
	}
}

func TestFindModrinthVersion_ReturnsFirstMatch(t *testing.T) {
	// Edge case: duplicate versions (shouldn't happen in practice)
	versions := []ModrinthVersion{
		{VersionNumber: "1.0.0", Files: []ModrinthFile{{URL: "first"}}},
		{VersionNumber: "1.0.0", Files: []ModrinthFile{{URL: "second"}}},
	}

	result := findModrinthVersion(versions, "1.0.0")

	if result == nil {
		t.Fatal("expected to find version")
	}
	if len(result.Files) == 0 || result.Files[0].URL != "first" {
		t.Error("expected first matching version to be returned")
	}
}

// --- findModrinthPrimaryFile() Tests ---

func TestFindModrinthPrimaryFile_FindsPrimaryFile(t *testing.T) {
	files := []ModrinthFile{
		{URL: "https://example.com/secondary.jar", Primary: false},
		{URL: "https://example.com/primary.jar", Primary: true},
		{URL: "https://example.com/another.jar", Primary: false},
	}

	result := findModrinthPrimaryFile(files)

	if result == nil {
		t.Fatal("expected to find primary file")
	}
	if result.URL != "https://example.com/primary.jar" {
		t.Errorf("expected URL %q, got %q", "https://example.com/primary.jar", result.URL)
	}
}

func TestFindModrinthPrimaryFile_ReturnsNilWhenNoPrimary(t *testing.T) {
	files := []ModrinthFile{
		{URL: "https://example.com/file1.jar", Primary: false},
		{URL: "https://example.com/file2.jar", Primary: false},
	}

	result := findModrinthPrimaryFile(files)

	if result != nil {
		t.Errorf("expected nil when no primary file, got %v", result)
	}
}

func TestFindModrinthPrimaryFile_ReturnsNilForEmptyArray(t *testing.T) {
	files := []ModrinthFile{}

	result := findModrinthPrimaryFile(files)

	if result != nil {
		t.Errorf("expected nil for empty array, got %v", result)
	}
}

func TestFindModrinthPrimaryFile_ReturnsNilForNilSlice(t *testing.T) {
	var files []ModrinthFile = nil

	result := findModrinthPrimaryFile(files)

	if result != nil {
		t.Errorf("expected nil for nil slice, got %v", result)
	}
}

func TestFindModrinthPrimaryFile_ReturnsFirstPrimary(t *testing.T) {
	// Edge case: multiple primary files (shouldn't happen in practice)
	files := []ModrinthFile{
		{URL: "https://example.com/first.jar", Primary: true},
		{URL: "https://example.com/second.jar", Primary: true},
	}

	result := findModrinthPrimaryFile(files)

	if result == nil {
		t.Fatal("expected to find primary file")
	}
	if result.URL != "https://example.com/first.jar" {
		t.Error("expected first primary file to be returned")
	}
}

func TestFindModrinthPrimaryFile_PreservesFileData(t *testing.T) {
	files := []ModrinthFile{
		{
			URL:     "https://example.com/plugin.jar",
			Primary: true,
			Hashes: struct {
				Sha512 string `json:"sha512"`
			}{
				Sha512: "abc123hash",
			},
		},
	}

	result := findModrinthPrimaryFile(files)

	if result == nil {
		t.Fatal("expected to find primary file")
	}
	if result.Hashes.Sha512 != "abc123hash" {
		t.Errorf("expected hash %q, got %q", "abc123hash", result.Hashes.Sha512)
	}
}

// --- CustomPluginSource.GetPluginDownload() Tests ---

func TestCustomPluginSource_GetPluginDownload_ValidConfig(t *testing.T) {
	source := &CustomPluginSource{}
	downloadURL := "https://example.com/plugin.jar"
	pluginConfig := config.PluginConfig{
		Source:      config.PluginSourceCustom,
		DownloadURL: &downloadURL,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if download.URL != downloadURL {
		t.Errorf("expected URL %q, got %q", downloadURL, download.URL)
	}
	if download.Checksum != "nocheck" {
		t.Errorf("expected checksum %q, got %q", "nocheck", download.Checksum)
	}
	if download.ChecksumType != ChecksumTypeSha256 {
		t.Errorf("expected checksum type %q, got %q", ChecksumTypeSha256, download.ChecksumType)
	}
}

func TestCustomPluginSource_GetPluginDownload_MissingURL(t *testing.T) {
	source := &CustomPluginSource{}
	pluginConfig := config.PluginConfig{
		Source:      config.PluginSourceCustom,
		DownloadURL: nil,
	}

	_, err := source.GetPluginDownload(pluginConfig)

	if err == nil {
		t.Fatal("expected error for missing download URL")
	}
}

func TestCustomPluginSource_GetPluginDownload_EmptyVersion(t *testing.T) {
	source := &CustomPluginSource{}
	downloadURL := "https://example.com/plugin.jar"
	pluginConfig := config.PluginConfig{
		Source:      config.PluginSourceCustom,
		DownloadURL: &downloadURL,
		Version:     nil,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Version should be empty for custom plugins (no version tracking)
	if download.Version != "" {
		t.Errorf("expected empty version, got %q", download.Version)
	}
}

func TestCustomPluginSource_GetPluginDownload_EmptyStringURL(t *testing.T) {
	// Edge case: URL is not nil but is empty string
	source := &CustomPluginSource{}
	emptyURL := ""
	pluginConfig := config.PluginConfig{
		Source:      config.PluginSourceCustom,
		DownloadURL: &emptyURL,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	// Current behavior: accepts empty URL (potential issue, but documenting current behavior)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if download.URL != "" {
		t.Errorf("expected empty URL, got %q", download.URL)
	}
}

// =============================================================================
// Network Tests - These hit real APIs
// =============================================================================

// --- Modrinth API Tests ---

func TestModrinthPluginSource_GetPluginDownload_LatestVersion(t *testing.T) {
	source := &ModrinthPluginSource{
		apiURL: "https://api.modrinth.com/v2",
	}
	resource := "chunky"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourceModrinth,
		Resource: &resource,
		Version:  nil, // Get latest
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("failed to get plugin download: %v", err)
	}

	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}
	if download.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
	if download.ChecksumType != ChecksumTypeSha512 {
		t.Errorf("expected checksum type %q, got %q", ChecksumTypeSha512, download.ChecksumType)
	}
	if download.Version == "" {
		t.Error("expected non-empty version")
	}

	t.Logf("Modrinth latest: version=%s, url=%s", download.Version, download.URL)
}

func TestModrinthPluginSource_GetPluginDownload_PinnedVersion(t *testing.T) {
	source := &ModrinthPluginSource{
		apiURL: "https://api.modrinth.com/v2",
	}
	resource := "chunky"
	version := "1.4.16"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourceModrinth,
		Resource: &resource,
		Version:  &version,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("failed to get plugin download: %v", err)
	}

	if download.Version != version {
		t.Errorf("expected version %q, got %q", version, download.Version)
	}
	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}
	if download.Checksum == "" {
		t.Error("expected non-empty checksum")
	}

	t.Logf("Modrinth pinned: version=%s, url=%s", download.Version, download.URL)
}

func TestModrinthPluginSource_GetPluginDownload_NonexistentPlugin(t *testing.T) {
	source := &ModrinthPluginSource{
		apiURL: "https://api.modrinth.com/v2",
	}
	resource := "this-plugin-definitely-does-not-exist-12345"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourceModrinth,
		Resource: &resource,
		Version:  nil,
	}

	_, err := source.GetPluginDownload(pluginConfig)

	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestModrinthPluginSource_GetPluginDownload_NonexistentVersion(t *testing.T) {
	source := &ModrinthPluginSource{
		apiURL: "https://api.modrinth.com/v2",
	}
	resource := "chunky"
	version := "99.99.99"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourceModrinth,
		Resource: &resource,
		Version:  &version,
	}

	_, err := source.GetPluginDownload(pluginConfig)

	if err == nil {
		t.Error("expected error for nonexistent version")
	}
}

func TestModrinthPluginSource_GetPluginDownload_EmptyStringVersion(t *testing.T) {
	// Edge case: Version is pointer to empty string (treated as unpinned/latest)
	source := &ModrinthPluginSource{
		apiURL: "https://api.modrinth.com/v2",
	}
	resource := "chunky"
	emptyVersion := ""
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourceModrinth,
		Resource: &resource,
		Version:  &emptyVersion,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("failed to get plugin download: %v", err)
	}

	// Empty string version should be treated as unpinned (get latest)
	if download.Version == "" {
		t.Error("expected non-empty version (should get latest)")
	}

	t.Logf("Modrinth empty string version (treated as latest): version=%s", download.Version)
}

// --- Paper Hangar API Tests ---

func TestPaperHangarPluginSource_GetPluginDownload_LatestVersion(t *testing.T) {
	source := &PaperHangarPluginSource{
		apiURL: "https://hangar.papermc.io/api/v1",
	}
	resource := "ViaVersion"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourcePaperHangar,
		Resource: &resource,
		Version:  nil, // Get latest
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("failed to get plugin download: %v", err)
	}

	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}
	if download.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
	if download.ChecksumType != ChecksumTypeSha256 {
		t.Errorf("expected checksum type %q, got %q", ChecksumTypeSha256, download.ChecksumType)
	}
	if download.Version == "" {
		t.Error("expected non-empty version")
	}

	t.Logf("Hangar latest: version=%s, url=%s", download.Version, download.URL)
}

func TestPaperHangarPluginSource_GetPluginDownload_PinnedVersion(t *testing.T) {
	source := &PaperHangarPluginSource{
		apiURL: "https://hangar.papermc.io/api/v1",
	}
	resource := "ViaVersion"
	version := "5.0.3"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourcePaperHangar,
		Resource: &resource,
		Version:  &version,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("failed to get plugin download: %v", err)
	}

	if download.Version != version {
		t.Errorf("expected version %q, got %q", version, download.Version)
	}
	if download.URL == "" {
		t.Error("expected non-empty download URL")
	}
	if download.Checksum == "" {
		t.Error("expected non-empty checksum")
	}

	t.Logf("Hangar pinned: version=%s, url=%s", download.Version, download.URL)
}

func TestPaperHangarPluginSource_GetPluginDownload_NonexistentPlugin(t *testing.T) {
	source := &PaperHangarPluginSource{
		apiURL: "https://hangar.papermc.io/api/v1",
	}
	resource := "ThisPluginDefinitelyDoesNotExist12345"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourcePaperHangar,
		Resource: &resource,
		Version:  nil,
	}

	_, err := source.GetPluginDownload(pluginConfig)

	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestPaperHangarPluginSource_GetPluginDownload_NonexistentVersion(t *testing.T) {
	source := &PaperHangarPluginSource{
		apiURL: "https://hangar.papermc.io/api/v1",
	}
	resource := "ViaVersion"
	version := "99.99.99"
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourcePaperHangar,
		Resource: &resource,
		Version:  &version,
	}

	_, err := source.GetPluginDownload(pluginConfig)

	if err == nil {
		t.Error("expected error for nonexistent version")
	}
}

func TestPaperHangarPluginSource_GetPluginDownload_EmptyStringVersion(t *testing.T) {
	// Edge case: Version is pointer to empty string (treated as unpinned/latest)
	source := &PaperHangarPluginSource{
		apiURL: "https://hangar.papermc.io/api/v1",
	}
	resource := "ViaVersion"
	emptyVersion := ""
	pluginConfig := config.PluginConfig{
		Source:   config.PluginSourcePaperHangar,
		Resource: &resource,
		Version:  &emptyVersion,
	}

	download, err := source.GetPluginDownload(pluginConfig)

	if err != nil {
		t.Fatalf("failed to get plugin download: %v", err)
	}

	// Empty string version should be treated as unpinned (get latest)
	if download.Version == "" {
		t.Error("expected non-empty version (should get latest)")
	}

	t.Logf("Hangar empty string version (treated as latest): version=%s", download.Version)
}

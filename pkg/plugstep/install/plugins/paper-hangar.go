package plugins

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

type PaperHangarPluginSource struct {
	apiURL string
}

type PaperHangarVersion struct {
	Downloads map[string]PaperHangarDownload `json:"downloads"`
}

type PaperHangarDownload struct {
	FileInfo struct {
		Sha256Hash string `json:"sha256Hash"`
	} `json:"fileInfo"`
	DownloadUrl string `json:"DownloadUrl"`
}

func (m *PaperHangarPluginSource) GetPluginDownload(c config.PluginConfig) (*PluginDownload, error) {
	isPinned := c.Version != nil && *c.Version != ""
	cache := GetCache()

	version := ""
	if isPinned {
		version = *c.Version
	} else {
		latest, err := m.getLatestVersion(*c.Resource)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version: %w", err)
		}
		version = latest
	}

	// Check permanent cache for resolved download
	downloadCacheKey := fmt.Sprintf("hangar:%s:%s:download", *c.Resource, version)
	var cached PluginDownload
	if cache != nil && cache.Get(downloadCacheKey, &cached) {
		return &cached, nil
	}

	url := fmt.Sprintf("%s/projects/%s/versions/%s", m.apiURL, *c.Resource, version)
	r, err := utils.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("got %d from %s", r.StatusCode, url)
	}

	var response PaperHangarVersion
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	downloadInfo, ok := response.Downloads["PAPER"]
	if !ok {
		return nil, fmt.Errorf("download not found on version")
	}

	download := &PluginDownload{
		URL:          downloadInfo.DownloadUrl,
		Checksum:     downloadInfo.FileInfo.Sha256Hash,
		ChecksumType: ChecksumTypeSha256,
		Version:      version,
	}

	// Cache permanently for this specific version
	if cache != nil {
		cache.SetPermanent(downloadCacheKey, download)
	}

	return download, nil
}

func (m *PaperHangarPluginSource) getLatestVersion(resource string) (string, error) {
	cacheKey := fmt.Sprintf("hangar:%s:latest", resource)

	var version string
	if cache := GetCache(); cache != nil && cache.Get(cacheKey, &version) {
		return version, nil
	}

	url := fmt.Sprintf("%s/projects/%s/latestrelease", m.apiURL, resource)
	r, err := utils.HTTPClient.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return "", fmt.Errorf("got %d from %s", r.StatusCode, url)
	}

	// Response is plain text, not JSON
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	version = strings.TrimSpace(string(body))

	if cache := GetCache(); cache != nil {
		cache.Set(cacheKey, version)
	}

	return version, nil
}

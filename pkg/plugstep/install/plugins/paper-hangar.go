package plugins

import (
	"encoding/json"
	"fmt"

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
	version := ""
	if c.Version != nil && *c.Version != "" {
		version = *c.Version
	} else {
		latest, err := m.getLatestVersion(*c.Resource)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest version: %w", err)
		}
		version = latest
	}

	cacheKey := fmt.Sprintf("hangar:%s:%s", *c.Resource, version)

	var response PaperHangarVersion
	if cache := GetCache(); cache != nil && cache.Get(cacheKey, &response) {
		// Cache hit
	} else {
		url := fmt.Sprintf("%s/projects/%s/versions/%s", m.apiURL, *c.Resource, version)
		r, err := utils.HTTPClient.Get(url)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		if r.StatusCode != 200 {
			return nil, fmt.Errorf("got %d from %s", r.StatusCode, url)
		}

		err = json.NewDecoder(r.Body).Decode(&response)
		if err != nil {
			return nil, err
		}

		if cache := GetCache(); cache != nil {
			cache.Set(cacheKey, response)
		}
	}

	download, ok := response.Downloads["PAPER"]
	if !ok {
		return nil, fmt.Errorf("download not found on version")
	}

	return &PluginDownload{
		URL:          download.DownloadUrl,
		Checksum:     download.FileInfo.Sha256Hash,
		ChecksumType: ChecksumTypeSha256,
	}, nil
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

	if err := json.NewDecoder(r.Body).Decode(&version); err != nil {
		return "", err
	}

	if cache := GetCache(); cache != nil {
		cache.Set(cacheKey, version)
	}

	return version, nil
}

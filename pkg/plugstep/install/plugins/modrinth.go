package plugins

import (
	"encoding/json"
	"fmt"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

type ModrinthPluginSource struct {
	apiURL string
}

type ModrinthVersion struct {
	VersionNumber string         `json:"version_number"`
	Files         []ModrinthFile `json:"files"`
}

type ModrinthFile struct {
	Hashes struct {
		Sha512 string `json:"sha512"`
	} `json:"hashes"`
	URL     string `json:"url"`
	Primary bool   `json:"primary"`
}

func (m *ModrinthPluginSource) GetPluginDownload(c config.PluginConfig) (*PluginDownload, error) {
	isPinned := c.Version != nil && *c.Version != ""
	cache := GetCache()

	// For pinned versions, check permanent cache first
	if isPinned {
		downloadCacheKey := fmt.Sprintf("modrinth:%s:%s:download", *c.Resource, *c.Version)
		var cached PluginDownload
		if cache != nil && cache.Get(downloadCacheKey, &cached) {
			return &cached, nil
		}
	}

	// Fetch version list (short TTL cache for latest discovery)
	versionsCacheKey := fmt.Sprintf("modrinth:%s:versions", *c.Resource)
	var response []ModrinthVersion
	if cache != nil && cache.Get(versionsCacheKey, &response) {
		// Cache hit
	} else {
		url := fmt.Sprintf("%s/project/%s/version", m.apiURL, *c.Resource)
		r, err := utils.HTTPClient.Get(url)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		if r.StatusCode != 200 {
			return nil, fmt.Errorf("got %d", r.StatusCode)
		}

		err = json.NewDecoder(r.Body).Decode(&response)
		if err != nil {
			return nil, err
		}

		if cache != nil {
			cache.Set(versionsCacheKey, response) // Short TTL
		}
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("no versions found for plugin")
	}

	var version *ModrinthVersion
	if isPinned {
		version = findModrinthVersion(response, *c.Version)
		if version == nil {
			return nil, fmt.Errorf("plugin version not found: %s", *c.Version)
		}
	} else {
		version = &response[0]
	}

	file := findModrinthPrimaryFile(version.Files)
	if file == nil {
		return nil, fmt.Errorf("plugin version has no primary file")
	}

	download := &PluginDownload{
		URL:          file.URL,
		Checksum:     file.Hashes.Sha512,
		ChecksumType: ChecksumTypeSha512,
	}

	// Cache permanently for this specific version
	if cache != nil {
		downloadCacheKey := fmt.Sprintf("modrinth:%s:%s:download", *c.Resource, version.VersionNumber)
		cache.SetPermanent(downloadCacheKey, download)
	}

	return download, nil
}

func findModrinthVersion(response []ModrinthVersion, version string) *ModrinthVersion {
	for _, resp := range response {
		if resp.VersionNumber == version {
			return &resp
		}
	}
	return nil
}

func findModrinthPrimaryFile(files []ModrinthFile) *ModrinthFile {
	for _, f := range files {
		if f.Primary == true {
			return &f
		}
	}
	return nil
}

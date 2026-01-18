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
	url := fmt.Sprintf("%s/projects/%s/versions/%s", m.apiURL, *c.Resource, *c.Version)
	r, err := utils.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("got %d, sent to %d", r.StatusCode, url)
	}

	var response PaperHangarVersion
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
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

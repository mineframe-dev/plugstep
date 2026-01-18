package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/pernydev/plugstep/pkg/plugstep/config"
)

type ServerJarVendor interface {
	GetDownload(config config.ServerConfig) (*ServerJarDownload, error)
}

type ServerJarDownload struct {
	URL      string
	Checksum string
}

type PaperJarVendor struct {
	apiURL string
}

func (p *PaperJarVendor) GetDownload(config config.ServerConfig) (*ServerJarDownload, error) {
	r, err := http.Get(fmt.Sprintf("%s/v3/projects/%s/versions/%s/builds/%s", p.apiURL, config.Project, config.MinecraftVersion, config.Version))
	if err != nil {
		return nil, err
	}

	var response struct {
		Downloads map[string]struct {
			Url       string `json:"url"`
			Checksums struct {
				Sha256 string `json:"sha256"`
			} `json:"checksums"`
		} `json:"downloads"`
	}

	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	download, ok := response.Downloads["server:default"]
	if !ok {
		log.Error("no server download avaliable for version")
		return nil, fmt.Errorf("no server download avaliable for version")
	}

	jar := ServerJarDownload{
		URL:      download.Url,
		Checksum: download.Checksums.Sha256,
	}

	return &jar, nil
}

func GetVendor(vendor config.ServerJarVendor) (ServerJarVendor, error) {
	switch vendor {
	case config.ServerJarVendorPaperMC:
		return &PaperJarVendor{
			apiURL: "https://fill.papermc.io",
		}, nil
	}
	return nil, fmt.Errorf("unknown server vendor: %s", vendor)
}

package server

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

const cacheName = "server"

type ServerJarVendor interface {
	GetDownload(config config.ServerConfig) (*ServerJarDownload, error)
}

type ServerJarDownload struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

type PaperJarVendor struct {
	apiURL string
}

func initCache() *utils.Cache {
	return utils.InitCache(cacheName)
}

func (p *PaperJarVendor) GetDownload(cfg config.ServerConfig) (*ServerJarDownload, error) {
	cache := initCache()
	cacheKey := fmt.Sprintf("paper:%s:%s:%s", cfg.Project, cfg.MinecraftVersion, cfg.Version)

	var cached ServerJarDownload
	if cache != nil && cache.Get(cacheKey, &cached) {
		return &cached, nil
	}

	r, err := utils.HTTPClient.Get(fmt.Sprintf("%s/v3/projects/%s/versions/%s/builds/%s", p.apiURL, cfg.Project, cfg.MinecraftVersion, cfg.Version))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

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

	if cache != nil {
		cache.SetPermanent(cacheKey, jar)
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

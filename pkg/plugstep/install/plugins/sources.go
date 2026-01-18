package plugins

import (
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
)

type PluginSource interface {
	GetPluginDownload(c config.PluginConfig) (*PluginDownload, error)
}

type ChecksumType string

const (
	ChecksumTypeSha256 ChecksumType = "sha256"
	ChecksumTypeSha512 ChecksumType = "sha512"
)

type PluginDownload struct {
	URL          string
	Checksum     string
	ChecksumType ChecksumType
}

func GetSource(source config.PluginSource) PluginSource {
	switch source {
	case config.PluginSourceModrinth:
		return &ModrinthPluginSource{
			apiURL: "https://api.modrinth.com/v2",
		}
	case config.PluginSourcePaperHangar:
		return &PaperHangarPluginSource{
			apiURL: "https://hangar.papermc.io/api/v1",
		}
	case config.PluginSourceCustom:
		return &CustomPluginSource{}
	}
	return nil
}

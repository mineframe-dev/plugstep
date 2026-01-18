package plugins

import (
	"fmt"

	"github.com/pernydev/plugstep/pkg/plugstep/config"
)

type CustomPluginSource struct{}

func (source *CustomPluginSource) GetPluginDownload(c config.PluginConfig) (*PluginDownload, error) {
	if c.DownloadURL == nil {
		return nil, fmt.Errorf("download URL is required for custom plugin source")
	}
	return &PluginDownload{
		URL:          *c.DownloadURL,
		Checksum:     "nocheck",
		ChecksumType: ChecksumTypeSha256,
	}, nil
}

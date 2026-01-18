package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
)

func LoadPlugstepConfig(configLocation string) (*PlugstepConfig, error) {
	var config PlugstepConfig
	log.Debug("loading config", "configLocation", configLocation)

	if _, err := toml.DecodeFile(configLocation, &config); err != nil {
		return nil, err
	}

	if config.Server == (ServerConfig{}) {
		return nil, fmt.Errorf("failed to find server config value in %s", configLocation)
	}

	return &config, nil
}

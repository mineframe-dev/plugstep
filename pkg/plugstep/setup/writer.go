package setup

import (
	"os"

	"github.com/BurntSushi/toml"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
)

func WriteConfig(path string, result *SetupResult) error {
	cfg := config.PlugstepConfig{
		Server: config.ServerConfig{
			Vendor:           config.ServerJarVendor(result.Vendor),
			Project:          result.Project,
			MinecraftVersion: result.MinecraftVersion,
			Version:          result.BuildVersion,
		},
		Plugins: []config.PluginConfig{},
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

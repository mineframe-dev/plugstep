package plugstep

import (
	"path/filepath"

	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
)

type Plugstep struct {
	Args            []string
	ServerDirectory string
	Config          *config.PlugstepConfig
}

func (p *Plugstep) Init() error {
	c, err := config.LoadPlugstepConfig(filepath.Join(p.ServerDirectory, "plugstep.toml"))
	if err != nil {
		log.Error("Failed to load Plugstep config", "err", err)
		return err
	}
	p.Config = c
	return nil
}

func CreatePlugstep(args []string, serverDirectory string) *Plugstep {
	return &Plugstep{
		Args:            args,
		ServerDirectory: serverDirectory,
	}
}

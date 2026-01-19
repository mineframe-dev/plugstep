package plugstep

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/setup"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

type Plugstep struct {
	Args            []string
	ServerDirectory string
	Config          *config.PlugstepConfig
}

func (p *Plugstep) Init() error {
	// Initialize cache database
	if err := utils.InitCacheDB(p.ServerDirectory); err != nil {
		log.Debug("Failed to initialize cache", "err", err)
	}

	configPath := filepath.Join(p.ServerDirectory, "plugstep.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := p.runSetupWizard(configPath); err != nil {
			return err
		}
	}

	c, err := config.LoadPlugstepConfig(configPath)
	if err != nil {
		log.Error("Failed to load Plugstep config", "err", err)
		return err
	}
	p.Config = c
	return nil
}

func (p *Plugstep) runSetupWizard(configPath string) error {
	log.Info("No plugstep.toml found. Starting interactive setup...")

	wizard := setup.NewSetupWizard()
	result, err := wizard.Run()
	if err != nil {
		log.Error("Setup wizard failed", "err", err)
		return err
	}

	if err := setup.WriteConfig(configPath, result); err != nil {
		log.Error("Failed to write config", "err", err)
		return err
	}

	log.Info("Created plugstep.toml successfully!")
	return nil
}

func CreatePlugstep(args []string, serverDirectory string) *Plugstep {
	return &Plugstep{
		Args:            args,
		ServerDirectory: serverDirectory,
	}
}

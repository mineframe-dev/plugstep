package commands

import (
	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/install/plugins"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/install/server"
)

func InstallCommand(ps *plugstep.Plugstep) {
	log.Debug("Installing server JAR and all plugins...", "serverjar", ps.Config.Server.Project, "minecraft-version", ps.Config.Server.MinecraftVersion, "plugins", len(ps.Config.Plugins))
	server.InstallServer(ps)
	plugins.InstallPlugins(ps)
}

package server

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/pernydev/plugstep/pkg/plugstep"
	"github.com/pernydev/plugstep/pkg/plugstep/utils"
)

func InstallServer(ps *plugstep.Plugstep) {
	var box = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#bac2de")).
		PaddingLeft(4).
		PaddingRight(4).
		Bold(true).
		Border(lipgloss.MarkdownBorder())

	var key = lipgloss.NewStyle().
		Bold(true).
		Width(30)

	var val = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7f849c"))

	b := box.Render(
		key.Render("Server config:") + "\n\n" +
			key.Render("Server vendor") + val.Render(string(ps.Config.Server.Vendor)) + "\n" +
			key.Render("Server project") + val.Render(ps.Config.Server.Project) + "\n" +
			key.Render("Server Minecraft version") + val.Render(ps.Config.Server.MinecraftVersion) + "\n" +
			key.Render("Server version") + val.Render(ps.Config.Server.Version),
	)
	fmt.Println(b)

	vendor, err := GetVendor(ps.Config.Server.Vendor)
	if err != nil {
		log.Error("failed to get server vendor", "err", err)
		return
	}
	download, err := vendor.GetDownload(ps.Config.Server)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debug("download found", "url", download.URL, "checksum", download.Checksum)

	location := filepath.Join(ps.ServerDirectory, "server.jar")

	existingJarChecksum, err := utils.CalculateFileSHA256(location)
	if err != nil {
		log.Debug("failed to get checksum of current serverjar", "err", err)
	}

	if existingJarChecksum == download.Checksum {
		log.Info("Checked server jar.")
		return
	}

	log.Info("Downloading server jar....")
	err = utils.DownloadFile(download.URL, location)
	if err != nil {
		log.Error("failed to download server jar", "err", err)
		return
	}

	log.Info("Downloaded server JAR successfully.")
}

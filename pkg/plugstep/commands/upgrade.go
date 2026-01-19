package commands

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
	"github.com/charmbracelet/log"
)

const latestVersionURL = "https://releases.perny.dev/mineframe/plugstep/latest"

func UpgradeCommand(serverDirectory string, targetVersion string) {
	versionFile := filepath.Join(serverDirectory, ".plugstep-version")

	currentVersion := ""
	if data, err := os.ReadFile(versionFile); err == nil {
		currentVersion = strings.TrimSpace(string(data))
	}

	var newVersion string
	if targetVersion != "" {
		newVersion = targetVersion
	} else {
		r, err := utils.HTTPClient.Get(latestVersionURL)
		if err != nil {
			log.Error("Failed to fetch latest version", "err", err)
			return
		}
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("Failed to read latest version", "err", err)
			return
		}

		newVersion = strings.TrimSpace(string(body))
	}

	if currentVersion == newVersion {
		log.Info("Already on version", "version", currentVersion)
		return
	}

	if err := os.WriteFile(versionFile, []byte(newVersion), 0644); err != nil {
		log.Error("Failed to write version file", "err", err)
		return
	}

	if currentVersion == "" {
		log.Info("Set version", "version", newVersion)
	} else {
		log.Info("Upgraded Plugstep Wrapper pinned version", "from", currentVersion, "to", newVersion)
	}
}

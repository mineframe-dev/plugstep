package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

const maxConcurrentDownloads = 4

var (
	statusLines = map[string]int{}
	renderMu    sync.Mutex
)

type installResult struct {
	plugin *config.PluginConfig
	status PluginInstallStatus
	err    error
}

func InstallPlugins(ps *plugstep.Plugstep) error {
	log.Info("Starting plugin download", "plugins", len(ps.Config.Plugins))

	InitCache()
	utils.EnsureDirectory(filepath.Join(ps.ServerDirectory, "plugins"))

	// Render all plugins as waiting
	for _, plugin := range ps.Config.Plugins {
		renderInstallBadge(&plugin, PluginInstallWaiting)
	}

	// Semaphore to limit concurrent downloads
	sem := make(chan struct{}, maxConcurrentDownloads)
	results := make(chan installResult, len(ps.Config.Plugins))

	var wg sync.WaitGroup
	for i := range ps.Config.Plugins {
		plugin := &ps.Config.Plugins[i]
		wg.Add(1)
		go func(p *config.PluginConfig) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			status, err := installPlugin(ps, p)
			results <- installResult{plugin: p, status: status, err: err}
		}(plugin)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	installed := 0
	checked := 0
	var errors []string
	pluginCount := len(ps.Config.Plugins)

	for result := range results {
		if result.err != nil {
			renderInstallBadge(result.plugin, PluginInstallFailed)
			errors = append(errors, fmt.Sprintf("%s: %v", *result.plugin.Resource, result.err))
		} else {
			renderInstallBadge(result.plugin, result.status)
			if result.status == PluginInstallStatusInstalled {
				installed++
			} else {
				checked++
			}
		}
		// Move cursor back to below all plugin lines
		fmt.Print(strings.Repeat("\033[E", pluginCount))
	}

	fmt.Printf("\r")
	fmt.Println("")

	if len(errors) > 0 {
		for _, e := range errors {
			log.Error("Plugin installation failed", "error", e)
		}
		return fmt.Errorf("%d plugin(s) failed to install", len(errors))
	}

	removed := removeOld(ps)

	log.Info("Plugins ready.", "installed", installed, "checked", checked, "removed", removed)

	return nil
}

// TODO: Add error handling
func removeOld(ps *plugstep.Plugstep) int {
	entries, err := os.ReadDir(filepath.Join(ps.ServerDirectory, "plugins"))
	if err != nil {
		log.Error("Error reading directory", "err", err)
		return 0
	}

	removed := 0

	for _, f := range entries {
		if f.IsDir() {
			continue
		}
		found := false
		for _, p := range ps.Config.Plugins {
			if f.Name() == *p.Resource+".jar" {
				found = true
				continue
			}
		}
		if found == true {
			continue
		}

		err := os.Remove(filepath.Join(ps.ServerDirectory, "plugins", f.Name()))
		if err != nil {
			log.Warn("failed to remove old plugin", "file", f.Name(), "err", err)
			continue
		}
		log.Infof("Removed %s", strings.Split(f.Name(), ".")[0])
		removed++
	}

	return removed
}

type PluginInstallStatus string

const (
	PluginInstallStatusInstalled PluginInstallStatus = "installed"
	PluginInstallStatusChecked   PluginInstallStatus = "checked"
	PluginInstallPrepairing      PluginInstallStatus = "prepairing"
	PluginInstallFailed          PluginInstallStatus = "failed"
	PluginInstallWaiting         PluginInstallStatus = "waiting"
)

func renderInstallBadge(p *config.PluginConfig, status PluginInstallStatus) {
	renderMu.Lock()
	defer renderMu.Unlock()

	if _, ok := statusLines[*p.Resource]; !ok {
		for k, v := range statusLines {
			statusLines[k] = v + 1
		}
		fmt.Println("")
		statusLines[*p.Resource] = 0
	}

	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4")).
		PaddingRight(1).
		Bold(true).
		PaddingLeft(1).
		Render("🡆")

	sourceBadge := lipgloss.NewStyle().
		Background(lipgloss.Color("#89b4fa")).
		Foreground(lipgloss.Color("#11111b")).
		PaddingRight(2).
		PaddingLeft(2).
		Transform(strings.ToUpper).
		Render(string(p.Source))

	background := "#f9e2af"
	switch status {
	case PluginInstallFailed:
		background = "#f38ba8"
	case PluginInstallStatusChecked:
		background = "#89dceb"
	case PluginInstallStatusInstalled:
		background = "#a6e3a1"
	}

	statusBadge := lipgloss.NewStyle().
		Background(lipgloss.Color(background)).
		Foreground(lipgloss.Color("#232634")).
		PaddingRight(2).
		PaddingLeft(2).
		Transform(strings.ToUpper).
		Render(string(status))

	cursorNav := strings.Repeat("\033[E", 999)

	fmt.Print(cursorNav + strings.Repeat("\033[F", statusLines[*p.Resource]))
	fmt.Printf("\r\033[K%s%s%s %s", badge, sourceBadge, statusBadge, *p.Resource)
}

func installPlugin(ps *plugstep.Plugstep, p *config.PluginConfig) (PluginInstallStatus, error) {
	source := GetSource(p.Source)
	if source == nil {
		return "", fmt.Errorf("invalid source")
	}

	download, err := source.GetPluginDownload(*p)
	if err != nil {
		return "", err
	}

	file := filepath.Join(ps.ServerDirectory, "plugins", *p.Resource+".jar")

	var hash string
	var hashErr error
	switch download.ChecksumType {
	case ChecksumTypeSha256:
		hash, hashErr = utils.CalculateFileSHA256(file)
	case ChecksumTypeSha512:
		hash, hashErr = utils.CalculateFileSHA512(file)
	}

	if hashErr != nil && !os.IsNotExist(hashErr) {
		log.Debug("failed to calculate file hash", "file", file, "err", hashErr)
	}

	if hash == download.Checksum {
		return PluginInstallStatusChecked, nil
	}

	err = utils.DownloadFile(download.URL, file)
	if err != nil {
		return PluginInstallFailed, fmt.Errorf("failed to download plugin: %w", err)
	}

	return PluginInstallStatusInstalled, nil
}

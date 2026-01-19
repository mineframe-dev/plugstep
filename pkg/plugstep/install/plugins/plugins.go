package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

const maxConcurrentDownloads = 25

type installResult struct {
	plugin *config.PluginConfig
	status PluginInstallStatus
	err    error
}

type PluginInstallStatus string

const (
	PluginInstallStatusInstalled PluginInstallStatus = "installed"
	PluginInstallStatusChecked   PluginInstallStatus = "checked"
	PluginInstallPreparing       PluginInstallStatus = "preparing"
	PluginInstallFailed          PluginInstallStatus = "failed"
	PluginInstallWaiting         PluginInstallStatus = "waiting"
)

type model struct {
	plugins   []pluginState
	results   <-chan installResult
	installed int
	checked   int
	errors    []string
	total     int
	done      bool
}

type pluginState struct {
	name   string
	source config.PluginSource
	status PluginInstallStatus
}

type pluginResultMsg installResult
type allDoneMsg struct{}

func (m model) Init() tea.Cmd {
	return waitForResult(m.results)
}

func waitForResult(results <-chan installResult) tea.Cmd {
	return func() tea.Msg {
		result, ok := <-results
		if !ok {
			return allDoneMsg{}
		}
		return pluginResultMsg(result)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pluginResultMsg:
		for i := range m.plugins {
			if m.plugins[i].name == *msg.plugin.Resource {
				if msg.err != nil {
					m.plugins[i].status = PluginInstallFailed
					m.errors = append(m.errors, fmt.Sprintf("%s: %v", *msg.plugin.Resource, msg.err))
				} else {
					m.plugins[i].status = msg.status
					if msg.status == PluginInstallStatusInstalled {
						m.installed++
					} else if msg.status == PluginInstallStatusChecked {
						m.checked++
					}
				}
				break
			}
		}
		return m, waitForResult(m.results)

	case allDoneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	for _, p := range m.plugins {
		b.WriteString(renderPluginLine(p))
		b.WriteString("\n")
	}

	return b.String()
}

func renderPluginLine(p pluginState) string {
	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4")).
		PaddingRight(1).
		Bold(true).
		PaddingLeft(1).
		Render("~")

	sourceBadge := lipgloss.NewStyle().
		Background(lipgloss.Color("#89b4fa")).
		Foreground(lipgloss.Color("#11111b")).
		PaddingRight(2).
		PaddingLeft(2).
		Transform(strings.ToUpper).
		Render(string(p.source))

	background := "#f9e2af"
	switch p.status {
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
		Render(string(p.status))

	return fmt.Sprintf("%s%s%s %s", badge, sourceBadge, statusBadge, p.name)
}

func InstallPlugins(ps *plugstep.Plugstep) error {
	log.Info("Starting plugin download", "plugins", len(ps.Config.Plugins))

	InitCache()
	utils.EnsureDirectory(filepath.Join(ps.ServerDirectory, "plugins"))

	plugins := make([]pluginState, len(ps.Config.Plugins))
	for i, p := range ps.Config.Plugins {
		plugins[i] = pluginState{
			name:   *p.Resource,
			source: p.Source,
			status: PluginInstallWaiting,
		}
	}

	sem := make(chan struct{}, maxConcurrentDownloads)
	results := make(chan installResult, len(ps.Config.Plugins))

	var wg sync.WaitGroup
	for i := range ps.Config.Plugins {
		plugin := &ps.Config.Plugins[i]
		wg.Add(1)
		go func(p *config.PluginConfig) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			status, err := installPlugin(ps, p)
			results <- installResult{plugin: p, status: status, err: err}
		}(plugin)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	m := model{
		plugins: plugins,
		results: results,
		total:   len(ps.Config.Plugins),
	}

	finalModel, err := tea.NewProgram(m).Run()
	if err != nil {
		return fmt.Errorf("UI error: %w", err)
	}

	fm := finalModel.(model)

	if len(fm.errors) > 0 {
		for _, e := range fm.errors {
			log.Error("Plugin installation failed", "error", e)
		}
		return fmt.Errorf("%d plugin(s) failed to install", len(fm.errors))
	}

	removed := removeOld(ps)

	log.Info("Plugins ready.", "installed", fm.installed, "checked", fm.checked, "removed", removed)

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

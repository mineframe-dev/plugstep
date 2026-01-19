package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/config"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/install/plugins"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

var (
	arrowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")).
			PaddingRight(1).
			Bold(true).
			PaddingLeft(1)

	modrinthBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("#1bd96a")).
			Foreground(lipgloss.Color("#11111b")).
			PaddingRight(1).
			PaddingLeft(1).
			Bold(true)

	hangarBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("#89b4fa")).
			Foreground(lipgloss.Color("#11111b")).
			PaddingRight(1).
			PaddingLeft(1).
			Bold(true)

	customBadge = lipgloss.NewStyle().
			Background(lipgloss.Color("#cba6f7")).
			Foreground(lipgloss.Color("#11111b")).
			PaddingRight(1).
			PaddingLeft(1).
			Bold(true)

	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4")).
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6adc8"))

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a6e3a1"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)
)

func PluginCommand(args []string, serverDirectory string) {
	if len(args) < 1 {
		showPluginHelp()
		return
	}

	switch args[0] {
	case "install", "i":
		pluginInstall(args[1:], serverDirectory)
	case "remove", "rm":
		pluginRemove(args[1:], serverDirectory)
	case "list", "ls":
		pluginList(serverDirectory)
	case "search", "s":
		pluginSearch(args[1:])
	default:
		log.Error("Unknown plugin subcommand", "subcommand", args[0])
		showPluginHelp()
	}
}

func showPluginHelp() {
	fmt.Println("Usage: plugstep plugin <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install, i  [source:name[@version]]  Install a plugin (interactive if no args)")
	fmt.Println("  remove, rm  <name>                   Remove a plugin")
	fmt.Println("  list, ls                             List installed plugins")
	fmt.Println("  search, s   <query>                  Search for plugins")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  plugstep plugin install                              (interactive)")
	fmt.Println("  plugstep plugin install modrinth:luckperms")
	fmt.Println("  plugstep plugin install hangar:FastAsyncWorldEdit@2.8.1")
	fmt.Println("  plugstep plugin remove luckperms")
	fmt.Println("  plugstep plugin search worldedit")
}

type PluginSpec struct {
	Source  string
	Name    string
	Version string
}

func parsePluginSpec(spec string) (*PluginSpec, error) {
	if !strings.Contains(spec, ":") {
		return nil, fmt.Errorf("invalid format: expected source:name[@version]")
	}

	parts := strings.SplitN(spec, ":", 2)
	source := parts[0]
	remainder := parts[1]

	var name, version string
	if strings.Contains(remainder, "@") {
		atParts := strings.SplitN(remainder, "@", 2)
		name = atParts[0]
		version = atParts[1]
	} else {
		name = remainder
	}

	validSources := map[string]bool{
		"modrinth": true,
		"hangar":   true,
		"custom":   true,
	}

	if !validSources[source] {
		return nil, fmt.Errorf("invalid source: %s (valid: modrinth, hangar, custom)", source)
	}

	if name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}

	return &PluginSpec{
		Source:  source,
		Name:    name,
		Version: version,
	}, nil
}

func loadConfig(serverDirectory string) (*config.PlugstepConfig, string, error) {
	configPath := filepath.Join(serverDirectory, "plugstep.toml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, configPath, fmt.Errorf("plugstep.toml not found - run 'plugstep' first to create config")
	}

	cfg, err := config.LoadPlugstepConfig(configPath)
	if err != nil {
		return nil, configPath, err
	}

	return cfg, configPath, nil
}

func saveConfig(configPath string, cfg *config.PlugstepConfig) error {
	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

func sourceToConfigSource(source string) config.PluginSource {
	switch source {
	case "modrinth":
		return config.PluginSourceModrinth
	case "hangar":
		return config.PluginSourcePaperHangar
	case "custom":
		return config.PluginSourceCustom
	}
	return ""
}

func pluginInstall(args []string, serverDirectory string) {
	var spec *PluginSpec
	var err error

	if len(args) < 1 {
		spec, err = interactivePluginInstall()
		if err != nil {
			log.Error("Interactive install failed", "err", err)
			return
		}
		if spec == nil {
			return
		}
	} else {
		spec, err = parsePluginSpec(args[0])
		if err != nil {
			log.Error("Invalid plugin spec", "err", err)
			return
		}
	}

	cfg, configPath, err := loadConfig(serverDirectory)
	if err != nil {
		log.Error("Failed to load config", "err", err)
		return
	}

	for _, p := range cfg.Plugins {
		if p.Resource != nil && *p.Resource == spec.Name {
			log.Warn("Plugin already configured", "name", spec.Name)
			return
		}
	}

	newPlugin := config.PluginConfig{
		Source:   sourceToConfigSource(spec.Source),
		Resource: &spec.Name,
	}
	if spec.Version != "" {
		newPlugin.Version = &spec.Version
	}

	log.Info("Validating plugin...", "source", spec.Source, "name", spec.Name)
	source := plugins.GetSource(newPlugin.Source)
	if source == nil {
		log.Error("Invalid plugin source", "source", spec.Source)
		return
	}

	_, err = source.GetPluginDownload(newPlugin)
	if err != nil {
		log.Error("Plugin or version not found", "name", spec.Name, "err", err)
		return
	}

	cfg.Plugins = append(cfg.Plugins, newPlugin)

	if err := saveConfig(configPath, cfg); err != nil {
		log.Error("Failed to save config", "err", err)
		return
	}

	log.Info("Added plugin to config", "source", spec.Source, "name", spec.Name)

	ps := &plugstep.Plugstep{
		ServerDirectory: serverDirectory,
		Config:          cfg,
	}
	plugins.InstallPlugins(ps)
}

func pluginList(serverDirectory string) {
	cfg, _, err := loadConfig(serverDirectory)
	if err != nil {
		log.Error("Failed to load config", "err", err)
		return
	}

	if len(cfg.Plugins) == 0 {
		fmt.Println(descStyle.Render("No plugins configured"))
		return
	}

	fmt.Println(headerStyle.Render(fmt.Sprintf("PLUGINS (%d)", len(cfg.Plugins))))
	for _, p := range cfg.Plugins {
		name := ""
		if p.Resource != nil {
			name = *p.Resource
		}
		version := "(latest)"
		if p.Version != nil && *p.Version != "" {
			version = *p.Version
		}

		badge := getSourceBadge(string(p.Source))
		fmt.Printf("  %s %s %s %s\n",
			arrowStyle.Render("→"),
			badge,
			nameStyle.Render(name),
			versionStyle.Render(version),
		)
	}
}

func getSourceBadge(source string) string {
	switch source {
	case "modrinth":
		return modrinthBadge.Render("MODRINTH")
	case "paper-hangar":
		return hangarBadge.Render("HANGAR")
	case "custom":
		return customBadge.Render("CUSTOM")
	default:
		return customBadge.Render(strings.ToUpper(source))
	}
}

func interactivePluginInstall() (*PluginSpec, error) {
	var source string
	sourceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select plugin source").
				Options(
					huh.NewOption("Modrinth", "modrinth"),
					huh.NewOption("Hangar (PaperMC)", "hangar"),
				).
				Value(&source),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := sourceForm.Run(); err != nil {
		return nil, err
	}

	var query string
	searchForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Search for a plugin").
				Placeholder("e.g. worldedit, luckperms").
				Value(&query),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := searchForm.Run(); err != nil {
		return nil, err
	}

	if query == "" {
		log.Warn("No search query provided")
		return nil, nil
	}

	log.Info("Searching for plugins...")

	var results []searchResult
	switch source {
	case "modrinth":
		results = searchModrinth(query)
	case "hangar":
		results = searchHangar(query)
	}

	if len(results) == 0 {
		log.Warn("No plugins found")
		return nil, nil
	}

	pluginOptions := make([]huh.Option[string], len(results))
	for i, r := range results {
		desc := r.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		label := fmt.Sprintf("%s - %s", r.Name, desc)
		pluginOptions[i] = huh.NewOption(label, r.Name)
	}

	var selectedPlugin string
	pluginForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a plugin to install").
				Options(pluginOptions...).
				Height(10).
				Value(&selectedPlugin),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := pluginForm.Run(); err != nil {
		return nil, err
	}

	var useLatest bool
	versionChoiceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use latest version?").
				Description("Recommended for staying up-to-date").
				Value(&useLatest),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := versionChoiceForm.Run(); err != nil {
		return nil, err
	}

	spec := &PluginSpec{
		Source: source,
		Name:   selectedPlugin,
	}

	if !useLatest {
		var version string
		versionForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter version").
					Placeholder("e.g. 5.4.0, 2.8.1").
					Value(&version),
			),
		).WithTheme(huh.ThemeCatppuccin())

		if err := versionForm.Run(); err != nil {
			return nil, err
		}
		spec.Version = version
	}

	return spec, nil
}

func pluginRemove(args []string, serverDirectory string) {
	if len(args) < 1 {
		log.Error("Usage: plugstep plugin remove <name>")
		return
	}

	name := args[0]

	cfg, configPath, err := loadConfig(serverDirectory)
	if err != nil {
		log.Error("Failed to load config", "err", err)
		return
	}

	found := false
	newPlugins := make([]config.PluginConfig, 0, len(cfg.Plugins))
	for _, p := range cfg.Plugins {
		if p.Resource != nil && *p.Resource == name {
			found = true
			continue
		}
		newPlugins = append(newPlugins, p)
	}

	if !found {
		log.Warn("Plugin not found in config", "name", name)
		return
	}

	cfg.Plugins = newPlugins

	if err := saveConfig(configPath, cfg); err != nil {
		log.Error("Failed to save config", "err", err)
		return
	}

	pluginsDir := filepath.Join(serverDirectory, "plugins")
	files, err := os.ReadDir(pluginsDir)
	if err == nil {
		for _, f := range files {
			if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(name)) && strings.HasSuffix(f.Name(), ".jar") {
				jarPath := filepath.Join(pluginsDir, f.Name())
				if err := os.Remove(jarPath); err == nil {
					log.Info("Deleted plugin file", "file", f.Name())
				}
			}
		}
	}

	log.Info("Removed plugin from config", "name", name)
}

type searchResult struct {
	Source      string
	Name        string
	Description string
}

func pluginSearch(args []string) {
	if len(args) < 1 {
		log.Error("Usage: plugstep plugin search <query>")
		return
	}

	query := strings.Join(args, " ")

	modrinthResults := searchModrinth(query)
	hangarResults := searchHangar(query)

	if len(modrinthResults) == 0 && len(hangarResults) == 0 {
		fmt.Println(descStyle.Render("No plugins found"))
		return
	}

	if len(modrinthResults) > 0 {
		fmt.Println(headerStyle.Render(fmt.Sprintf("MODRINTH (%d)", len(modrinthResults))))
		for _, r := range modrinthResults {
			desc := r.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("  %s %s %s\n",
				arrowStyle.Render("→"),
				nameStyle.Width(25).Render(r.Name),
				descStyle.Render(desc),
			)
		}
	}

	if len(hangarResults) > 0 {
		fmt.Println(headerStyle.Render(fmt.Sprintf("HANGAR (%d)", len(hangarResults))))
		for _, r := range hangarResults {
			desc := r.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Printf("  %s %s %s\n",
				arrowStyle.Render("→"),
				nameStyle.Width(25).Render(r.Name),
				descStyle.Render(desc),
			)
		}
	}
}

func searchModrinth(query string) []searchResult {
	searchURL := fmt.Sprintf(
		"https://api.modrinth.com/v2/search?query=%s&facets=%s&limit=10",
		url.QueryEscape(query),
		url.QueryEscape(`[["project_type:plugin"]]`),
	)

	r, err := utils.HTTPClient.Get(searchURL)
	if err != nil {
		log.Debug("Modrinth search failed", "err", err)
		return nil
	}
	defer r.Body.Close()

	var response struct {
		Hits []struct {
			Slug        string `json:"slug"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		log.Debug("Modrinth search parse failed", "err", err)
		return nil
	}

	var results []searchResult
	for _, hit := range response.Hits {
		results = append(results, searchResult{
			Source:      "modrinth",
			Name:        hit.Slug,
			Description: hit.Description,
		})
	}

	return results
}

func searchHangar(query string) []searchResult {
	searchURL := fmt.Sprintf(
		"https://hangar.papermc.io/api/v1/projects?q=%s&limit=10",
		url.QueryEscape(query),
	)

	r, err := utils.HTTPClient.Get(searchURL)
	if err != nil {
		log.Debug("Hangar search failed", "err", err)
		return nil
	}
	defer r.Body.Close()

	var response struct {
		Result []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		log.Debug("Hangar search parse failed", "err", err)
		return nil
	}

	var results []searchResult
	for _, p := range response.Result {
		results = append(results, searchResult{
			Source:      "hangar",
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return results
}

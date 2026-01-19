package setup

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
)

type SetupResult struct {
	Vendor           string
	Project          string
	MinecraftVersion string
	BuildVersion     string
}

type SetupWizard struct {
	papermc *PaperMCClient
}

func NewSetupWizard() *SetupWizard {
	return &SetupWizard{
		papermc: NewPaperMCClient(),
	}
}

func (w *SetupWizard) Run() (*SetupResult, error) {
	result := &SetupResult{}

	vendorForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select server vendor").
				Options(
					huh.NewOption("PaperMC", "papermc"),
				).
				Value(&result.Vendor),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := vendorForm.Run(); err != nil {
		return nil, fmt.Errorf("vendor selection failed: %w", err)
	}

	log.Info("Fetching available projects...")
	projects, err := w.papermc.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	projectOptions := make([]huh.Option[string], len(projects))
	for i, p := range projects {
		projectOptions[i] = huh.NewOption(p.Name, p.ID)
	}

	projectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select project").
				Options(projectOptions...).
				Value(&result.Project),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := projectForm.Run(); err != nil {
		return nil, fmt.Errorf("project selection failed: %w", err)
	}

	log.Info("Fetching available versions...")
	versions, err := w.papermc.GetVersions(result.Project)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}

	versionOptions := make([]huh.Option[string], len(versions))
	for i, v := range versions {
		versionOptions[i] = huh.NewOption(v, v)
	}

	versionForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Minecraft version").
				Options(versionOptions...).
				Height(10).
				Value(&result.MinecraftVersion),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := versionForm.Run(); err != nil {
		return nil, fmt.Errorf("version selection failed: %w", err)
	}

	var useLatest bool
	buildChoiceForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Use latest build?").
				Description("Recommended for staying up-to-date").
				Value(&useLatest),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := buildChoiceForm.Run(); err != nil {
		return nil, fmt.Errorf("build choice failed: %w", err)
	}

	if useLatest {
		result.BuildVersion = "latest"
	} else {
		log.Info("Fetching available builds...")
		builds, err := w.papermc.GetBuilds(result.Project, result.MinecraftVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch builds: %w", err)
		}

		buildOptions := make([]huh.Option[string], len(builds))
		for i, b := range builds {
			buildOptions[i] = huh.NewOption(b, b)
		}

		buildForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select build number").
					Options(buildOptions...).
					Height(10).
					Value(&result.BuildVersion),
			),
		).WithTheme(huh.ThemeCatppuccin())

		if err := buildForm.Run(); err != nil {
			return nil, fmt.Errorf("build selection failed: %w", err)
		}
	}

	return result, nil
}

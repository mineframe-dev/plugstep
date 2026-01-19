package setup

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
)

func parseVersion(v string) ([]int, string) {
	parts := strings.SplitN(v, "-", 2)
	base := parts[0]
	suffix := ""
	if len(parts) > 1 {
		suffix = parts[1]
	}

	baseParts := strings.Split(base, ".")
	nums := make([]int, len(baseParts))
	for i, p := range baseParts {
		n, _ := strconv.Atoi(p)
		nums[i] = n
	}
	return nums, suffix
}

func compareVersions(a, b string) int {
	partsA, suffixA := parseVersion(a)
	partsB, suffixB := parseVersion(b)

	maxLen := max(len(partsA), len(partsB))
	for i := 0; i < maxLen; i++ {
		var valA, valB int
		if i < len(partsA) {
			valA = partsA[i]
		}
		if i < len(partsB) {
			valB = partsB[i]
		}
		if c := cmp.Compare(valA, valB); c != 0 {
			return c
		}
	}

	if suffixA == "" && suffixB != "" {
		return 1
	}
	if suffixA != "" && suffixB == "" {
		return -1
	}

	return cmp.Compare(suffixA, suffixB)
}

type PaperMCClient struct {
	baseURL string
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewPaperMCClient() *PaperMCClient {
	return &PaperMCClient{baseURL: "https://fill.papermc.io"}
}

func (c *PaperMCClient) GetProjects() ([]Project, error) {
	r, err := utils.HTTPClient.Get(fmt.Sprintf("%s/v3/projects", c.baseURL))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var response struct {
		Projects []struct {
			Project Project `json:"project"`
		} `json:"projects"`
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}

	projects := make([]Project, len(response.Projects))
	for i, p := range response.Projects {
		projects[i] = p.Project
	}

	return projects, nil
}

func (c *PaperMCClient) GetVersions(project string) ([]string, error) {
	r, err := utils.HTTPClient.Get(fmt.Sprintf("%s/v3/projects/%s", c.baseURL, project))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var response struct {
		Versions map[string][]string `json:"versions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}

	var versions []string
	for _, minorVersions := range response.Versions {
		versions = append(versions, minorVersions...)
	}

	slices.SortFunc(versions, func(a, b string) int {
		return compareVersions(b, a)
	})

	return versions, nil
}

func (c *PaperMCClient) GetBuilds(project, version string) ([]string, error) {
	r, err := utils.HTTPClient.Get(fmt.Sprintf("%s/v3/projects/%s/versions/%s/builds", c.baseURL, project, version))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var response struct {
		Builds []struct {
			Build int `json:"build"`
		} `json:"builds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, err
	}

	builds := make([]string, len(response.Builds))
	for i, b := range response.Builds {
		builds[len(response.Builds)-1-i] = fmt.Sprintf("%d", b.Build)
	}

	return builds, nil
}

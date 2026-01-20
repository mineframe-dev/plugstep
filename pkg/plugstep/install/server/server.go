package server

import (
	"fmt"
	"path/filepath"
	"strings"

	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type serverModel struct {
	downloaded int64
	total      int64
	done       bool
	err        error
	status     string
	progressCh <-chan progressUpdate
}

type progressUpdate struct {
	downloaded int64
	total      int64
}

type progressMsg progressUpdate
type doneMsg struct{ err error }

func (m serverModel) Init() tea.Cmd {
	return waitForProgress(m.progressCh)
}

func waitForProgress(ch <-chan progressUpdate) tea.Cmd {
	return func() tea.Msg {
		p, ok := <-ch
		if !ok {
			return doneMsg{}
		}
		return progressMsg(p)
	}
}

func (m serverModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressMsg:
		m.downloaded = msg.downloaded
		m.total = msg.total
		m.status = "downloading"
		return m, waitForProgress(m.progressCh)
	case doneMsg:
		m.done = true
		m.err = msg.err
		if m.err == nil {
			m.status = "installed"
		} else {
			m.status = "failed"
		}
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m serverModel) View() string {
	var b strings.Builder

	badge := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cdd6f4")).
		PaddingRight(1).
		Bold(true).
		PaddingLeft(1).
		Render("~")

	background := "#f9e2af"
	switch m.status {
	case "downloading":
		background = "#89b4fa"
	case "installed":
		background = "#a6e3a1"
	case "failed":
		background = "#f38ba8"
	case "checked":
		background = "#89dceb"
	}

	statusBadge := lipgloss.NewStyle().
		Background(lipgloss.Color(background)).
		Foreground(lipgloss.Color("#232634")).
		PaddingRight(2).
		PaddingLeft(2).
		Transform(strings.ToUpper).
		Render(m.status)

	b.WriteString(fmt.Sprintf("%s%s server.jar ", badge, statusBadge))

	if m.status == "downloading" {
		b.WriteString(renderProgressBar(m.downloaded, m.total, 20))
	}

	b.WriteString("\n")
	return b.String()
}

func renderProgressBar(downloaded, total int64, width int) string {
	var percent float64
	if total > 0 {
		percent = float64(downloaded) / float64(total)
	}

	filled := int(percent * float64(width))
	if filled > width {
		filled = width
	}

	filledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6e3a1"))
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#45475a"))

	bar := filledStyle.Render(strings.Repeat("—", filled)) +
		emptyStyle.Render(strings.Repeat("—", width-filled))

	return fmt.Sprintf("[%s] %s/%s", bar, formatSize(downloaded), formatSize(total))
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)
	switch {
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func InstallServer(ps *plugstep.Plugstep) {
	var box = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#bac2de")).
		PaddingLeft(4).
		PaddingRight(4).
		Bold(true).
		Border(lipgloss.RoundedBorder())

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

	progressCh := make(chan progressUpdate)
	doneCh := make(chan error)

	go func() {
		err := utils.DownloadFileWithProgress(download.URL, location, func(downloaded, total int64) {
			progressCh <- progressUpdate{downloaded: downloaded, total: total}
		})
		close(progressCh)
		doneCh <- err
	}()

	m := serverModel{
		status:     "preparing",
		progressCh: progressCh,
	}

	finalModel, teaErr := tea.NewProgram(m).Run()
	if teaErr != nil {
		log.Error("UI error", "err", teaErr)
		return
	}

	downloadErr := <-doneCh
	fm := finalModel.(serverModel)

	if downloadErr != nil || fm.err != nil {
		if downloadErr != nil {
			log.Error("failed to download server jar", "err", downloadErr)
		}
		return
	}

	log.Info("Downloaded server JAR successfully.")
}

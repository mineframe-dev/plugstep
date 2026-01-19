package main

import (
	_ "embed"
	"flag"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep"
	"forgejo.perny.dev/mineframe/plugstep/pkg/plugstep/commands"
)

//go:embed ascii.txt
var ascii string
var emoticon = "♪(๑ᴖ◡ᴖ๑)♪"

var Version = "dev"

var debug *bool
var serverDirectory *string

func init() {
	debug = flag.Bool("d", false, "enable debug logging")
	serverDirectory = flag.String("dir", ".", "path to server")
}

func main() {
	log.Info("Initializing Plugstep " + emoticon)

	flag.Parse()

	if debug != nil && *debug {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("Debug logging enabled.")

	args := flag.Args()
	log.Debug("Loaded args", "args", args)

	if len(args) < 1 {
		ShowVersion()
		return
	}
	command := args[0]

	switch command {
	case "version", "v":
		ShowVersion()
		return
	case "upgrade", "u":
		targetVersion := ""
		if len(args) > 1 {
			targetVersion = args[1]
		}
		commands.UpgradeCommand(*serverDirectory, targetVersion)
		return
	case "plugin", "p":
		commands.PluginCommand(args[1:], *serverDirectory)
		return
	}

	ps := plugstep.CreatePlugstep(args, *serverDirectory)
	err := ps.Init()
	if err != nil {
		return
	}

	switch command {
	case "install", "i":
		commands.InstallCommand(ps)
		return
	}

	log.Info("Unknown command", "command", command)
}

func ShowVersion() {
	var box = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#cba6f7")).
		PaddingLeft(4).
		PaddingRight(4).
		Bold(true).
		Border(lipgloss.MarkdownBorder())

	var copyright = lipgloss.NewStyle().Foreground(lipgloss.Color("#7f849c"))

	fmt.Println(box.Render(
		copyright.Render("Copyright © Perny and McWar team") + "\n" +
			ascii + "\n" +
			"Version: " + Version,
	))
}

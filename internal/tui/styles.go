package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	base     = lipgloss.Color("#1E1E2E")
	text     = lipgloss.Color("#CDD6F4")
	subtext0 = lipgloss.Color("#A6ADC8")
	mauve    = lipgloss.Color("#CBA6F7")
	pink     = lipgloss.Color("#F5C2E7")
	lavender = lipgloss.Color("#B4BEFE")
	green    = lipgloss.Color("#A6E3A1")
	red      = lipgloss.Color("#F38BA8")
	teal     = lipgloss.Color("#94E2D5")
	yellow   = lipgloss.Color("#F9E2AF")
	overlay0 = lipgloss.Color("#6C7086")
)

var (
	panelStyle  = lipgloss.NewStyle().Background(base).Padding(1, 2).BorderForeground(mauve).BorderStyle(lipgloss.RoundedBorder())
	titleStyle  = lipgloss.NewStyle().Foreground(mauve).Bold(true).Background(base)
	headerStyle = lipgloss.NewStyle().Foreground(lavender).Bold(true).Background(base)
	itemSel     = lipgloss.NewStyle().Foreground(pink).Bold(true).Background(base)
	itemUnsel   = lipgloss.NewStyle().Foreground(subtext0).Background(base)
	itemActive  = lipgloss.NewStyle().Foreground(teal).Background(base)
	itemOk      = lipgloss.NewStyle().Foreground(green).Background(base)
	itemErr     = lipgloss.NewStyle().Foreground(red).Background(base)
	itemSkip    = lipgloss.NewStyle().Foreground(yellow).Background(base)
	footerStyle = lipgloss.NewStyle().Foreground(overlay0).Background(base).Padding(0, 1)
)

func newSpinner() spinner.Model {
	return spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(teal).Background(base)),
	)
}

func bgLine(s string) string {
	return lipgloss.NewStyle().Background(base).Render(s)
}

func SetupLogger() *log.Logger {
	l := log.New(os.Stdout)
	l.SetReportTimestamp(false)
	return l
}

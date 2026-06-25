package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	catppuBase     = lipgloss.Color("#1E1E2E")
	catppuMauve    = lipgloss.Color("#CBA6F7")
	catppuPink     = lipgloss.Color("#F5C2E7")
	catppuLavender = lipgloss.Color("#B4BEFE")
	catppuGreen    = lipgloss.Color("#A6E3A1")
	catppuRed      = lipgloss.Color("#F38BA8")
	catppuTeal     = lipgloss.Color("#94E2D5")
	catppuYellow   = lipgloss.Color("#F9E2AF")
	catppuOverlay  = lipgloss.Color("#6C7086")
	catppuText     = lipgloss.Color("#CDD6F4")
	catppuSub      = lipgloss.Color("#A6ADC8")
)

var (
	panelStyle = lipgloss.NewStyle().
			Background(catppuBase).
			Padding(1, 2).
			Margin(0, 1).
			BorderForeground(catppuMauve).
			BorderStyle(lipgloss.RoundedBorder())

	titleStyle = lipgloss.NewStyle().
			Foreground(catppuMauve).
			Bold(true).
			Background(catppuBase)

	headerStyle = lipgloss.NewStyle().
			Foreground(catppuLavender).
			Bold(true).
			Background(catppuBase)

	itemSel = lipgloss.NewStyle().
		Foreground(catppuPink).
		Bold(true).
		Background(catppuBase)

	itemUnsel = lipgloss.NewStyle().
			Foreground(catppuSub).
			Background(catppuBase)

	itemActive = lipgloss.NewStyle().
			Foreground(catppuTeal).
			Background(catppuBase)

	itemOk = lipgloss.NewStyle().
		Foreground(catppuGreen).
		Background(catppuBase)

	itemErr = lipgloss.NewStyle().
		Foreground(catppuRed).
		Background(catppuBase)

	itemSkip = lipgloss.NewStyle().
			Foreground(catppuYellow).
			Background(catppuBase)

	footerStyle = lipgloss.NewStyle().
			Foreground(catppuOverlay).
			Background(catppuBase).
			Padding(0, 1)

	mauveStyle = lipgloss.NewStyle().
			Foreground(catppuMauve).
			Background(catppuBase)
)

func newSpinner() spinner.Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(catppuTeal).Background(catppuBase)),
	)
	return s
}

func SetupLogger() *log.Logger {
	l := log.New(os.Stdout)
	l.SetReportTimestamp(false)
	return l
}

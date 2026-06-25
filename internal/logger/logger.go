package logger

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	catppuBase   = lipgloss.Color("#1E1E2E")
	catppuMauve  = lipgloss.Color("#CBA6F7")
	catppuPeach  = lipgloss.Color("#FFCDD6")
	catppuTeal   = lipgloss.Color("#94E2D5")
	catppuYellow = lipgloss.Color("#F9E2AF")
	catppuRed    = lipgloss.Color("#F38BA8")
	catppuText   = lipgloss.Color("#CDD6F4")
	catppuSub    = lipgloss.Color("#A6ADC8")
)

var s = log.DefaultStyles()

func init() {
	s.Levels[log.DebugLevel] = lipgloss.NewStyle().
		Padding(0, 1).
		Background(catppuMauve).
		Foreground(catppuBase).
		SetString(" D ")

	s.Levels[log.InfoLevel] = lipgloss.NewStyle().
		Padding(0, 1).
		Background(catppuTeal).
		Foreground(catppuBase).
		SetString(" I ")

	s.Levels[log.WarnLevel] = lipgloss.NewStyle().
		Padding(0, 1).
		Background(catppuYellow).
		Foreground(catppuBase).
		SetString(" W ")

	s.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		Padding(0, 1).
		Background(catppuRed).
		Foreground(catppuBase).
		SetString(" E ")

	s.Levels[log.FatalLevel] = lipgloss.NewStyle().
		Padding(0, 1).
		Background(catppuRed).
		Foreground(catppuBase).
		Bold(true).
		SetString(" F ")

	s.Timestamp = lipgloss.NewStyle().
		Foreground(catppuSub)

	s.Message = lipgloss.NewStyle().
		Foreground(catppuText)

	s.Key = lipgloss.NewStyle().
		Foreground(catppuMauve)

	s.Value = lipgloss.NewStyle().
		Foreground(catppuText)

}

func New() *log.Logger {
	l := log.New(os.Stderr)
	l.SetStyles(s)
	l.SetReportTimestamp(true)
	l.SetTimeFormat("15:04:05")
	l.SetLevel(log.InfoLevel)
	return l
}

func Debug() *log.Logger {
	l := New()
	l.SetLevel(log.DebugLevel)
	return l
}

func ForInstance(instType string) *log.Logger {
	l := New()
	bg := catppuMauve
	if strings.ToLower(instType) == "radarr" {
		bg = catppuPeach
	}
	badge := lipgloss.NewStyle().
		Foreground(catppuBase).
		Background(bg).
		Bold(true).
		Padding(0, 1).
		Render(strings.ToUpper(instType))
	l.SetPrefix(badge + " ")
	return l
}

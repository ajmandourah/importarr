package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"importarr/internal/api"
	"importarr/internal/models"
)

type phase int

const (
	phaseInstances phase = iota
	phaseQueue
	phaseImporting
	phaseResults
)

type App struct {
	width     int
	height    int
	instances []models.Instance
	selected  map[int]bool
	cursor    int
	phase     phase

	currentInstance models.Instance
	stuckRecords    []models.QueueRecord
	progress        ImportProgress
	spinner         spinner.Model
	spinnerStarted  bool
	fallback        bool
}

func New(instances []models.Instance, fallback bool) App {
	return App{
		instances: instances,
		selected:  make(map[int]bool),
		phase:     phaseInstances,
		fallback:  fallback,
		spinner:   newSpinner(),
	}
}

func (a App) Init() tea.Cmd {
	return spinner.Tick
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		switch a.phase {
		case phaseInstances:
			switch msg.String() {
			case "up", "k":
				if a.cursor > 0 {
					a.cursor--
				}
			case "down", "j":
				if a.cursor < len(a.instances)-1 {
					a.cursor++
				}
			case " ":
				a.selected[a.cursor] = !a.selected[a.cursor]
			case "enter":
				if countTrue(a.selected) > 0 {
					a.phase = phaseQueue
					a.cursor = 0
				}
			case "esc", "q":
				return a, tea.Quit
			}

		case phaseQueue:
			switch msg.String() {
			case "enter":
				a.phase = phaseImporting
				a.spinnerStarted = false
			case "esc", "q":
				a.phase = phaseInstances
			}

		case phaseImporting:
			switch msg.String() {
			case "esc", "q":
				return a, tea.Quit
			}

		case phaseResults:
			return a, tea.Quit
		}

	case spinner.TickMsg:
		a.spinner, cmd = a.spinner.Update(msg)
	}

	return a, cmd
}

func (a App) View() string {
	var content string

	switch a.phase {
	case phaseInstances:
		content = renderInstanceSelection(a.instances, a.selected, a.cursor)
	case phaseQueue:
		content = a.queueView()
	case phaseImporting:
		content = a.importingView()
	case phaseResults:
		content = renderResults(a.progress)
	}

	contentWidth := lipgloss.Width(content)
	contentHeight := lipgloss.Height(content)

	hPad := (a.width - contentWidth) / 2
	if hPad < 0 {
		hPad = 0
	}
	vPad := (a.height - contentHeight) / 2
	if vPad < 0 {
		vPad = 0
	}

	return lipgloss.NewStyle().
		Background(catppuBase).
		Width(a.width).
		Height(a.height).
		Padding(vPad, hPad).
		Render(content)
}

func (a *App) queueView() string {
	if a.currentInstance.Name == "" {
		return a.runImports()
	}
	return renderQueue(a.stuckRecords, a.currentInstance.Name)
}

func (a *App) importingView() string {
	if !a.spinnerStarted {
		a.spinnerStarted = true
		return a.runImports()
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("  Importing...") + "\n\n")
	b.WriteString(itemActive.Render("  "+a.spinner.View()+" "+a.progress.Phase) + "\n")
	if a.progress.PhaseDetail != "" {
		b.WriteString(itemUnsel.Render("  "+a.progress.PhaseDetail) + "\n")
	}
	return panelStyle.Render(b.String())
}

func (a *App) runImports() string {
	if a.currentInstance.Name == "" {
		for i := range a.selected {
			a.currentInstance = a.instances[i]
			break
		}
	}

	a.progress = ImportProgress{
		InstanceName: a.currentInstance.Name,
		Results:      make(map[string][]models.ImportResult),
		Errors:       make(map[string]string),
	}

	a.importInstance(a.currentInstance)

	if !a.hasNextInstance() {
		a.phase = phaseResults
	} else {
		a.nextInstance()
	}

	return renderResults(a.progress)
}

func (a *App) importInstance(inst models.Instance) {
	a.progress.Phase = "Fetching queue..."
	a.progress.PhaseDetail = inst.Name

	client, err := api.NewClient(inst)
	if err != nil {
		a.progress.Errors["_"] = err.Error()
		return
	}

	records, err := client.GetQueue()
	if err != nil {
		a.progress.Errors["_"] = err.Error()
		return
	}

	a.progress.Records = records
	a.progress.Phase = fmt.Sprintf("Processing %d stuck item(s)...", len(records))

	for _, record := range records {
		recKey := fmt.Sprintf("%d", record.ID)
		a.progress.PhaseDetail = record.Title

		files, err := client.GetManualImport(record)
		if err != nil {
			a.progress.Errors[recKey] = err.Error()
			if a.fallback {
				_ = client.RemoveFromQueue(record.ID)
				_ = client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber)
			}
			continue
		}

		if len(files) == 0 {
			a.progress.Errors[recKey] = "no importable files found"
			if a.fallback {
				_ = client.RemoveFromQueue(record.ID)
				_ = client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber)
			}
			continue
		}

		results, err := client.PostManualImport(files)
		if err != nil {
			a.progress.Errors[recKey] = err.Error()
			if a.fallback {
				_ = client.RemoveFromQueue(record.ID)
				_ = client.TriggerSearch(record.SeriesOrMovieID(), record.SeasonNumber)
			}
			continue
		}

		a.progress.Results[recKey] = results
	}

	a.progress.Phase = ""
	a.progress.PhaseDetail = ""
}

func (a *App) hasNextInstance() bool {
	for i, inst := range a.instances {
		if a.selected[i] && inst.Name != a.currentInstance.Name {
			return true
		}
	}
	return false
}

func (a *App) nextInstance() {
	for i, inst := range a.instances {
		if a.selected[i] && inst.Name != a.currentInstance.Name {
			a.currentInstance = inst
			a.phase = phaseQueue
			return
		}
	}
}

func Run(instances []models.Instance, fallback bool) {
	if len(instances) == 0 {
		fmt.Println("No instances configured. Add instances to .env or config.yaml.")
		os.Exit(1)
	}

	p := tea.NewProgram(New(instances, fallback), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

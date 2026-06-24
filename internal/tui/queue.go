package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"importarr/internal/models"
)

func renderQueue(records []models.QueueRecord, instanceName string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Stuck Queue Items ("+instanceName+")") + "\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %d item(s) found", len(records))) + "\n\n")

	if len(records) == 0 {
		b.WriteString(itemUnsel.Render("  No stuck items found.") + "\n")
		return panelStyle.Render(b.String())
	}

	colID := 6
	colTitle := 50
	colProtocol := 10
	colAge := 12

	b.WriteString(lipgloss.NewStyle().Background(base).Render(
		fmt.Sprintf("%-*s %-*s %-*s %s\n",
			colID, headerStyle.Render("ID"),
			colTitle, headerStyle.Render("Title"),
			colProtocol, headerStyle.Render("Protocol"),
			headerStyle.Render("Age"))))
	b.WriteString(lipgloss.NewStyle().Background(base).Render(
		fmt.Sprintf("%-*s %-*s %-*s %s\n",
			colID, strings.Repeat("-", colID),
			colTitle, strings.Repeat("-", colTitle),
			colProtocol, strings.Repeat("-", colProtocol),
			strings.Repeat("-", colAge))) + "\n")

	for _, r := range records {
		age := time.Since(r.Added)
		ageStr := formatDuration(age)
		title := r.Title
		if len(title) > colTitle {
			title = title[:colTitle-3] + "..."
		}

		b.WriteString(fmt.Sprintf("%-*s %-*s %-*s %s\n",
			colID, itemUnsel.Render(fmt.Sprintf("%d", r.ID)),
			colTitle, itemUnsel.Render(title),
			colProtocol, itemUnsel.Render(r.Protocol),
			itemUnsel.Render(ageStr)) + "\n")
	}

	b.WriteString("\n" + footerStyle.Render("  Press any key to continue"))
	return panelStyle.Render(b.String())
}

func formatDuration(d time.Duration) string {
	if d.Hours() >= 24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	return fmt.Sprintf("%.0fm", d.Minutes())
}

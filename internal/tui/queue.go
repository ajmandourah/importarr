package tui

import (
	"fmt"
	"strings"
	"time"

	"importarr/internal/models"
)

func renderQueue(records []models.QueueRecord, instanceName string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Stuck Queue Items ("+instanceName+")") + "\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %d item(s) found", len(records))) + "\n\n")

	if len(records) == 0 {
		b.WriteString(itemUnsel.Render("  No stuck items found.") + "\n")
		b.WriteString("\n" + footerStyle.Render("  Press Enter to continue"))
		return panelStyle.Render(b.String())
	}

	for i, r := range records {
		age := time.Since(r.Added)
		ageStr := formatDuration(age)

		title := r.Title
		b.WriteString("  " + mauveStyle.Render(fmt.Sprintf("#%d", r.ID)) + " " + mauveStyle.Render(title) + "\n")
		b.WriteString("  " + itemUnsel.Render("  "+r.Protocol+"  "+ageStr) + "\n")
		if i < len(records)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + footerStyle.Render("  Press Enter to import, Esc to exit"))
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

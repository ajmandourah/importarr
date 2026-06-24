package tui

import (
	"fmt"
	"strings"

	"importarr/internal/models"
)

type ImportProgress struct {
	InstanceName string
	Records      []models.QueueRecord
	Results      map[string][]models.ImportResult
	Errors       map[string]string
	Phase        string
	PhaseDetail  string
}

func renderResults(p ImportProgress) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Import Results ("+p.InstanceName+")") + "\n\n")

	if p.Phase != "" {
		b.WriteString(itemActive.Render("  "+p.Phase) + "\n")
		if p.PhaseDetail != "" {
			b.WriteString(itemUnsel.Render("  "+p.PhaseDetail) + "\n")
		}
		b.WriteString("\n")
	}

	totalOk, totalErr, totalSkip := 0, 0, 0

	for _, record := range p.Records {
		recKey := fmt.Sprintf("%d", record.ID)
		results, ok := p.Results[recKey]
		errMsg, hasErr := p.Errors[recKey]

		b.WriteString(headerStyle.Render("  "+record.Title) + "\n")

		if hasErr {
			b.WriteString("    " + itemErr.Render("[FAIL]") + " " + itemUnsel.Render(errMsg) + "\n")
			totalErr++
			continue
		}

		if !ok || len(results) == 0 {
			b.WriteString("    " + itemErr.Render("[FAIL]") + " " + itemUnsel.Render("no importable files found") + "\n")
			totalErr++
			continue
		}

		for _, r := range results {
			switch r.Status {
			case "imported":
				totalOk++
				b.WriteString("    " + itemOk.Render("[OK]") + " " + itemUnsel.Render(shortPath(r.Path)) + "\n")
			case "skipped":
				totalSkip++
				b.WriteString("    " + itemSkip.Render("[SKIP]") + " " + itemUnsel.Render(shortPath(r.Path)) + " " + itemUnsel.Render(r.Message) + "\n")
			case "rejected":
				totalErr++
				b.WriteString("    " + itemErr.Render("[REJECT]") + " " + itemUnsel.Render(shortPath(r.Path)) + " " + itemUnsel.Render(r.Message) + "\n")
			default:
				b.WriteString("    " + itemUnsel.Render("["+r.Status+"]") + " " + itemUnsel.Render(shortPath(r.Path)) + "\n")
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(headerStyle.Render("  Summary") + "\n")
	b.WriteString("    " + itemOk.Render(fmt.Sprintf("Imported: %d", totalOk)) + "\n")
	b.WriteString("    " + itemSkip.Render(fmt.Sprintf("Skipped: %d", totalSkip)) + "\n")
	b.WriteString("    " + itemErr.Render(fmt.Sprintf("Failed: %d", totalErr)) + "\n")

	b.WriteString("\n" + footerStyle.Render("  Press any key to exit"))
	return panelStyle.Render(b.String())
}

func shortPath(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) > 3 {
		return "..." + strings.Join(parts[len(parts)-3:], "/")
	}
	return p
}

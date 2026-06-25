package tui

import (
	"fmt"
	"strings"

	"importarr/internal/models"
)

func renderInstanceSelection(instances []models.Instance, selected map[int]bool, cursor int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Select Instances") + "\n\n")
	b.WriteString(headerStyle.Render("  Space: select  Enter: confirm  Esc: exit") + "\n\n")

	for i, inst := range instances {
		check := "[ ]"
		if selected[i] {
			check = "[x]"
		}

		name := inst.Name + " (" + inst.Type + ")"
		url := inst.URL

		if i == cursor {
			check = itemSel.Render(check)
			b.WriteString("  " + itemSel.Render("> ") + check + " " + itemSel.Render(name) + "\n")
			if url != "" {
				b.WriteString("  " + strings.Repeat(" ", len(check)+4) + itemUnsel.Render(url) + "\n")
			}
		} else {
			b.WriteString("  " + itemUnsel.Render("  ") + itemUnsel.Render(check) + " " + itemUnsel.Render(name) + "\n")
			if url != "" {
				b.WriteString("  " + strings.Repeat(" ", len(check)+4) + itemUnsel.Render(url) + "\n")
			}
		}
	}

	b.WriteString("\n" + footerStyle.Render(fmt.Sprintf("  %d selected", countTrue(selected))))
	return panelStyle.Render(b.String())
}

func countTrue(m map[int]bool) int {
	c := 0
	for _, v := range m {
		if v {
			c++
		}
	}
	return c
}

package tui

import (
	"fmt"
	"strings"

	"importarr/internal/models"
)

func renderInstances(instances []models.Instance, cursor int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Select Instances") + "\n\n")

	for i, inst := range instances {
		prefix := "  "
		name := fmt.Sprintf("%s (%s) - %s", inst.Name, inst.Type, inst.URL)
		if i == cursor {
			prefix = itemSel.Render("> ")
			b.WriteString(prefix + itemSel.Render(name) + "\n")
		} else {
			b.WriteString(prefix + itemUnsel.Render(name) + "\n")
		}
	}

	b.WriteString("\n" + footerStyle.Render("  Arrow keys: navigate  Enter: select  Esc: confirm"))
	return panelStyle.Render(b.String())
}

func renderInstanceSelection(instances []models.Instance, selected map[int]bool, cursor int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("  Select Instances") + "\n\n")
	b.WriteString(headerStyle.Render("  Press Enter to toggle, Esc to confirm") + "\n\n")

	for i, inst := range instances {
		prefix := "  "
		check := "[ ]"
		if selected[i] {
			check = "[x]"
		}
		if i == cursor {
			prefix = itemSel.Render("> ")
			check = itemSel.Render(check)
			b.WriteString(prefix + check + " " + itemSel.Render(inst.Name+" ("+inst.Type+")") + "\n")
		} else {
			b.WriteString(prefix + itemUnsel.Render(check) + " " + itemUnsel.Render(inst.Name+" ("+inst.Type+")") + "\n")
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

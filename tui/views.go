package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Intro-Shlok/AutoMate/core"
)

func domainColor(domain string) string {
	colors := map[string]string{
		"security":  colRed,
		"network":   colBlue,
		"system":    colGreen,
		"text":      colOrange,
		"dev":       colPurple,
		"container": colCyan,
	}
	if c, ok := colors[domain]; ok {
		return c
	}
	return colText
}

func domainCounts(tools []core.ToolDefinition) map[string]int {
	counts := make(map[string]int)
	for _, t := range tools {
		d := domainName(t)
		counts[d]++
	}
	return counts
}

func domainName(t core.ToolDefinition) string {
	if t.Namespace == "" {
		return "other"
	}
	for i := 0; i < len(t.Namespace); i++ {
		if t.Namespace[i] == ':' {
			return t.Namespace[:i]
		}
	}
	return t.Namespace
}

func toolsForDomain(tools []core.ToolDefinition, domain string) []core.ToolDefinition {
	if domain == "" || domain == "all" {
		return tools
	}
	var filtered []core.ToolDefinition
	for _, t := range tools {
		if domainName(t) == domain {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (m *Model) statusBarView() string {
	installed := 0
	for _, s := range m.statuses {
		if s.OnPath || s.DockerImage || s.PackageManager != "" {
			installed++
		}
	}

	hints := "  ↑↓ navigate  •  / search  •  ? help  •  q quit"
	stats := fmt.Sprintf(" %d tools  •  %d installed  •  v%s  ",
		len(m.tools), installed, appVersion)

	return lipgloss.JoinHorizontal(lipgloss.Bottom,
		statusBarLeft.Render(hints),
		statusBarRight.Render(stats),
	)
}

func (m *Model) sidebarView() string {
	var b strings.Builder

	b.WriteString(sideTitleStyle.Render(" Domains "))
	b.WriteString("\n\n")

	counts := domainCounts(m.tools)

	// "All" item
	allPrefix := "  "
	if m.selectedDomain == "" || m.selectedDomain == "all" {
		allPrefix = "▸ "
	}
	totalItems := len(m.tools)
	allLine := fmt.Sprintf("%s%-18s %d", allPrefix, "All", totalItems)
	if m.selectedDomain == "" || m.selectedDomain == "all" {
		b.WriteString(sideActiveStyle.Render(allLine))
	} else {
		b.WriteString(sideItemStyle.Render(allLine))
	}

	// Domain items
	for _, d := range core.Domains(m.tools) {
		prefix := "  "
		if m.selectedDomain == d {
			prefix = "▸ "
		}
		c := domainColor(d)
		count := counts[d]
		countStr := fmt.Sprintf("%d", count)
		line := fmt.Sprintf("%s%-18s %s", prefix, d, countStr)

		if m.selectedDomain == d {
			colored := lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Bold(true).Padding(0, 1).Render(line)
			b.WriteString(colored)
		} else {
			colored := lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Padding(0, 1).Render(line)
			b.WriteString(colored)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(itemDescStyle.Render("  ctrl+x b: hide"))
	b.WriteString("\n")

	return sideTitleStyle.Width(26).Render(" Domains ") + "\n" + b.String()
}

func (m *Model) bootView() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(logoStyle.Render(logo))

	b.WriteString(bootTitleStyle.Render(" System "))
	b.WriteString("\n\n")

	rows := []struct {
		key, val string
	}{
		{"Hostname", m.sysInfo.hostname},
		{"OS", m.sysInfo.os},
		{"Kernel", m.sysInfo.kernel},
		{"Architecture", m.sysInfo.arch},
		{"CPU Cores", fmt.Sprintf("%d", m.sysInfo.cpuCores)},
		{"Memory", m.sysInfo.memory},
		{"Uptime", m.sysInfo.uptime},
		{"Tools Loaded", fmt.Sprintf("%d", m.sysInfo.toolCount)},
		{"Version", m.sysInfo.buildVer},
		{"Started", m.sysInfo.timeStr},
	}

	for _, r := range rows {
		b.WriteString(fmt.Sprintf("  %-22s%s\n",
			infoKeyStyle.Render(r.key+":"),
			infoValStyle.Render(r.val),
		))
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colBorder)).Render("  " + strings.Repeat("─", 72)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Loading tools..."))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Press any key to skip"))

	return b.String()
}

func (m *Model) listView() string {
	var b strings.Builder

	// Title
	title := " AutoMate — Tool Manager "
	if m.selectedDomain != "" && m.selectedDomain != "all" {
		title = fmt.Sprintf(" AutoMate — %s ", m.selectedDomain)
	}
	b.WriteString(headerStyle.Render(title))

	// Sidebar + list in horizontal layout
	if m.sidebarVisible {
		b.WriteString("\n")
		side := m.sidebarView()
		listContent := m.toolListContent()
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, side, listContent))
	} else {
		b.WriteString("\n\n")
		b.WriteString(m.toolListContent())
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render(
		"  ↑↓ navigate  •  enter: details  •  /: search  •  i: install  •  r: run  •  t: terminal  •  ?: help  •  q: quit"))

	return b.String()
}

func (m *Model) toolListContent() string {
	var b strings.Builder

	// Filtered list header
	showing := len(m.filtered)
	total := len(m.tools)
	domain := ""
	if m.selectedDomain != "" && m.selectedDomain != "all" {
		domain = fmt.Sprintf(" [%s]", m.selectedDomain)
	}

	b.WriteString(statusBarLeft.Width(40).Background(lipgloss.Color(colBg)).Foreground(lipgloss.Color(colTextMuted)).
		Render(fmt.Sprintf("  Showing %d/%d tools%s", showing, total, domain)))
	b.WriteString("\n\n")

	for i, t := range m.filtered {
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
		}

		// Name column
		nameStr := prefix + t.Name
		if i == m.cursor {
			nameStr = itemSelectedStyle.Render(prefix + t.Name)
		} else {
			nameStr = itemStyle.Render(nameStr)
		}

		// Status indicator
		statusStr := "   "
		if s, ok := m.statuses[t.ID]; ok {
			if s.OnPath {
				statusStr = itemInstalledStyle.Render(" ✓ ")
			} else if s.PackageManager != "" || s.DockerImage {
				statusStr = itemInstalledStyle.Render(" ~ ")
			} else {
				statusStr = itemNotInstalledStyle.Render(" ✗ ")
			}
		}

		// Description
		desc := t.Description
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}

		b.WriteString(fmt.Sprintf("  %s %s %s\n", statusStr, nameStr, itemDescStyle.Render(desc)))
	}

	if len(m.filtered) == 0 {
		b.WriteString(itemDescStyle.Render("  No tools found."))
		b.WriteString("\n")
	}

	return b.String()
}

func (m *Model) detailView() string {
	if m.selected == nil {
		m.currentScreen = screenList
		return ""
	}

	t := m.selected
	var b strings.Builder

	b.WriteString(detailTitleStyle.Render(fmt.Sprintf(" %s ", t.Name)))

	// Info section
	b.WriteString("\n")
	infoRows := []struct {
		label, value string
	}{
		{"ID", t.ID},
		{"Namespace", t.Namespace},
		{"Risk", t.RiskLevel},
		{"Trust", t.TrustLevel},
	}
	for _, r := range infoRows {
		b.WriteString(fmt.Sprintf("\n  %s %s",
			detailLabelStyle.Render(r.label+":"),
			detailValueStyle.Render(r.value),
		))
	}

	// Description
	b.WriteString(fmt.Sprintf("\n\n  %s\n", detailValueStyle.Render(t.Description)))

	// Capabilities
	if len(t.Capabilities) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n    %s\n",
			detailSectionStyle.Render("Capabilities:"),
			detailValueStyle.Render(strings.Join(t.Capabilities, ", ")),
		))
	}

	// Install status
	if s, ok := m.statuses[t.ID]; ok {
		b.WriteString(fmt.Sprintf("\n  %s\n    ", detailSectionStyle.Render("Status:")))
		if s.OnPath {
			b.WriteString(itemInstalledStyle.Render(fmt.Sprintf("✓ Installed at %s", s.PathLocation)))
		} else if s.PackageManager != "" {
			b.WriteString(itemInstalledStyle.Render(fmt.Sprintf("~ Available via %s", s.PackageManager)))
		} else {
			b.WriteString(itemNotInstalledStyle.Render("✗ Not installed"))
		}
	}

	// Template
	if t.Execution.Template != "" {
		b.WriteString(fmt.Sprintf("\n\n  %s\n    %s\n",
			detailSectionStyle.Render("Template:"),
			detailValueStyle.Render(t.Execution.Template),
		))
	}

	// Parameters
	if len(t.Parameters) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", detailSectionStyle.Render(fmt.Sprintf("Parameters (%d):", len(t.Parameters)))))
		for _, p := range t.Parameters {
			req := " "
			if p.Required {
				req = "*"
			}
			def := ""
			if p.DefaultValue != nil {
				def = fmt.Sprintf(" [default: %v]", p.DefaultValue)
			}
			b.WriteString(fmt.Sprintf("    %s %s (%s)%s — %s\n", req, detailValueStyle.Render(p.Name), p.Type, def, p.Description))
		}
	}

	// Install methods
	if len(t.Install) > 0 {
		b.WriteString(fmt.Sprintf("\n  %s\n", detailSectionStyle.Render("Install methods:")))
		for _, inst := range t.Install {
			methodColor := colText
			switch inst.Method {
			case "apt":
				methodColor = colOrange
			case "pip":
				methodColor = colGreen
			case "go":
				methodColor = colBlue
			case "cargo":
				methodColor = colOrange
			case "brew":
				methodColor = colYellow
			case "docker":
				methodColor = colCyan
			case "git":
				methodColor = colPurple
			}
			methodStr := lipgloss.NewStyle().Foreground(lipgloss.Color(methodColor)).Render(inst.Method)
			pkg := ""
			if inst.PackageName != "" {
				pkg = fmt.Sprintf(" (%s)", inst.PackageName)
			}
			b.WriteString(fmt.Sprintf("    • %s%s\n", methodStr, pkg))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  i: install  •  r: run  •  enter: recheck  •  esc/q: back"))

	return detailStyle.Render(b.String()) + "\n"
}

func (m *Model) installView() string {
	if m.selected == nil {
		m.currentScreen = screenList
		return ""
	}

	var b strings.Builder
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf(" Install: %s ", m.selected.Name)))
	b.WriteString("\n\n")

	installMethods := m.selected.Install
	if len(installMethods) == 0 {
		b.WriteString("  No install methods defined.\n")
		b.WriteString("\n  Press any key to go back.\n")
		return b.String()
	}

	b.WriteString("  Available install methods:\n\n")
	for _, inst := range installMethods {
		methodColor := colText
		switch inst.Method {
		case "apt":
			methodColor = colOrange
		case "pip":
			methodColor = colGreen
		case "go":
			methodColor = colBlue
		case "cargo":
			methodColor = colOrange
		case "brew":
			methodColor = colYellow
		case "docker":
			methodColor = colCyan
		case "git":
			methodColor = colPurple
		}
		methodStr := lipgloss.NewStyle().Foreground(lipgloss.Color(methodColor)).Render(inst.Method)
		b.WriteString(fmt.Sprintf("    %s", methodStr))
		if inst.PackageName != "" {
			b.WriteString(fmt.Sprintf(" (%s)", inst.PackageName))
		}
		b.WriteString("\n")
		for _, c := range inst.Commands {
			b.WriteString(fmt.Sprintf("      → %s\n", c))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n  Install this tool? (y/N) ")
	b.WriteString("\n  Press 'y' to install, any other key to go back.\n")

	return installStyle.Render(b.String())
}

func (m *Model) runView() string {
	if m.selected == nil {
		m.currentScreen = screenList
		return ""
	}

	var b strings.Builder
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf(" Run: %s ", m.selected.Name)))
	b.WriteString("\n\n")

	if m.selected.Execution.Template == "" {
		b.WriteString("  No execution template defined.\n")
		return runStyle.Render(b.String())
	}

	b.WriteString(fmt.Sprintf("  Template: %s\n\n", m.selected.Execution.Template))

	if len(m.selected.Parameters) > 0 {
		b.WriteString("  Parameters:\n")
		for _, p := range m.selected.Parameters {
			req := ""
			if p.Required {
				req = " (required)"
			}
			b.WriteString(fmt.Sprintf("    %s: %s%s\n", p.Name, p.Type, req))
		}
	}

	b.WriteString("\n  Run this tool with default params? (y/N) ")
	b.WriteString("\n  Press 'y' to run, any other key to go back.\n")

	return runStyle.Render(b.String())
}

func (m *Model) searchView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" Search "))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("\n  %s█\n\n", searchStyle.Render(m.searchQuery)))

	// Live results
	if len(m.searchQuery) > 0 {
		results := core.SearchTools(m.tools, m.searchQuery)
		if len(results) > 0 {
			b.WriteString(fmt.Sprintf("  %d matches:\n", len(results)))
			for i, t := range results {
				if i > 9 {
					b.WriteString(fmt.Sprintf("  ... and %d more\n", len(results)-10))
					break
				}
				statusStr := " "
				if s, ok := m.statuses[t.ID]; ok {
					if s.OnPath {
						statusStr = "✓"
					} else {
						statusStr = "✗"
					}
				}
				b.WriteString(fmt.Sprintf("    %s %s\n", statusStr, t.Name))
			}
		} else {
			b.WriteString("  No matches.\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Type to search  •  enter: filter list  •  esc: cancel"))
	return b.String()
}

func (m *Model) terminalView() string {
	var b strings.Builder
	b.WriteString(detailTitleStyle.Render(" Terminal "))
	b.WriteString("\n\n")
	b.WriteString("  Open an interactive shell?\n\n")
	b.WriteString("  This will launch your default shell.\n")
	b.WriteString("  Type 'exit' to return to AutoMate.\n\n")
	b.WriteString("  Press any key to open shell, esc/q to go back.\n")
	return terminalStyle.Render(b.String())
}

func (m *Model) helpView() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render(" Help "))
	b.WriteString("\n\n")
	b.WriteString(helpHeaderStyle.Render("  Navigation"))
	b.WriteString("\n")
	b.WriteString("    ↑/↓ or j/k        — Move cursor\n")
	b.WriteString("    enter              — Select / open details\n")
	b.WriteString("    /                  — Search tools\n")
	b.WriteString("    q / ctrl+c         — Quit\n\n")
	b.WriteString(helpHeaderStyle.Render("  Actions"))
	b.WriteString("\n")
	b.WriteString("    i                  — Install selected tool\n")
	b.WriteString("    r                  — Run selected tool\n")
	b.WriteString("    t                  — Open interactive terminal\n")
	b.WriteString("    R                  — Refresh installation status\n\n")
	b.WriteString(helpHeaderStyle.Render("  Leader Key (ctrl+x)"))
	b.WriteString("\n")
	b.WriteString("    ctrl+x then b      — Toggle sidebar\n")
	b.WriteString("    ctrl+x then s      — System status view\n")
	b.WriteString("    ctrl+x then p      — Command palette\n\n")
	b.WriteString(helpHeaderStyle.Render("  Command Palette (ctrl+p)"))
	b.WriteString("\n")
	b.WriteString("    ctrl+p             — Open command palette\n")
	b.WriteString("    Type to fuzzy search tools by name/description\n")
	b.WriteString("    enter              — Open selected tool\n\n")
	b.WriteString(helpHeaderStyle.Render("  Screens"))
	b.WriteString("\n")
	b.WriteString("    esc / backspace    — Go back\n")
	b.WriteString("    ?                  — Toggle this help\n")
	b.WriteString("    ctrl+x n           — Reset to list view\n")

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Press any key to close help."))

	return b.String()
}

func (m *Model) paletteView() string {
	var b strings.Builder

	b.WriteString(paletteStyle.Render(""))
	// We need to build content manually because border + bg
	content := strings.Builder{}

	prompt := " Search: "
	if m.paletteQuery == "" {
		prompt += "█"
	} else {
		prompt += m.paletteQuery + "█"
	}

	content.WriteString(paletteInputStyle.Render(prompt))
	content.WriteString("\n\n")

	if len(m.paletteResults) > 0 {
		for i, t := range m.paletteResults {
			line := fmt.Sprintf("  %s  %s", t.Name, t.ID)
			if i == m.paletteCursor {
				content.WriteString(paletteActiveStyle.Render(line))
			} else {
				content.WriteString(paletteItemStyle.Render(line))
			}
			content.WriteString("\n")
		}
	} else if m.paletteQuery != "" {
		content.WriteString(paletteHintStyle.Render("  No matching tools found."))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(paletteHintStyle.Render("  ↑↓ navigate  •  enter: open  •  esc: close"))
	content.WriteString("\n")

	// Combine border with content
	b.WriteString("\n")
	b.WriteString(paletteStyle.Render(content.String()))

	return b.String()
}

func (m *Model) installProgressView() string {
	var b strings.Builder
	name := "unknown"
	if m.installingTool != nil {
		name = m.installingTool.Name
	}
	b.WriteString(detailTitleStyle.Render(fmt.Sprintf(" Installing: %s ", name)))
	b.WriteString("\n\n")

	if m.installOutput != "" {
		b.WriteString(fmt.Sprintf("  %s\n", m.installOutput))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  esc to cancel"))

	return installStyle.Render(b.String())
}



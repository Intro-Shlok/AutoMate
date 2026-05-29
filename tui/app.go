package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Intro-Shlok/AutoMate/core"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#22d3ee")).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#22d3ee")).
			BorderBottom(true).
			Width(80)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a1a1aa")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a855f7")).
			Bold(true).
			Padding(0, 1)

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10b981"))

	notInstalledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ef4444"))

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#22d3ee")).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#22d3ee")).
				BorderBottom(true).
				Width(60)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#636363")).
			Padding(0, 1)
)

type screen int

const (
	screenList screen = iota
	screenDetail
	screenInstall
	screenRun
	screenSearch
	screenTerminal
	screenHelp
)

type Model struct {
	tools          []core.ToolDefinition
	toolsCache     *core.Cache
	currentScreen  screen
	cursor         int
	filtered       []core.ToolDefinition
	selected       *core.ToolDefinition
	searchQuery    string
	statuses       map[string]core.InstallStatus
	detailContent  string
	helpVisible    bool
	width          int
	height         int
}

func NewApp(tools []core.ToolDefinition, cache *core.Cache) *Model {
	return &Model{
		tools:      tools,
		toolsCache: cache,
		filtered:   tools,
		statuses:   core.DetectAllTools(tools),
	}
}

func (m *Model) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.currentScreen {
	case screenList:
		return m.handleListKey(msg)
	case screenDetail:
		return m.handleDetailKey(msg)
	case screenInstall:
		return m.handleInstallKey(msg)
	case screenRun:
		return m.handleRunKey(msg)
	case screenSearch:
		return m.handleSearchKey(msg)
	case screenTerminal:
		return m.handleTerminalKey(msg)
	case screenHelp:
		if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
			m.currentScreen = screenList
		}
		return m, nil
	}
	return m, nil
}

func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.selected = &m.filtered[m.cursor]
			m.buildDetail()
			m.currentScreen = screenDetail
		}
	case "/":
		m.searchQuery = ""
		m.currentScreen = screenSearch
	case "i":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.selected = &m.filtered[m.cursor]
			m.currentScreen = screenInstall
		}
	case "r":
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.selected = &m.filtered[m.cursor]
			m.currentScreen = screenRun
		}
	case "t":
		m.currentScreen = screenTerminal
	case "?":
		m.currentScreen = screenHelp
	case "R":
		// Refresh detection status
		m.statuses = core.DetectAllTools(m.tools)
	}
	return m, nil
}

func (m *Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "backspace":
		m.currentScreen = screenList
	case "i":
		m.currentScreen = screenInstall
	case "r":
		m.currentScreen = screenRun
	case "enter":
		if m.selected != nil {
			status := core.DetectTool(*m.selected)
			m.statuses[m.selected.ID] = status
		}
	}
	return m, nil
}

func (m *Model) handleInstallKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.selected != nil {
			err := core.InstallTool(*m.selected, core.InstallOptions{})
			if err != nil {
				m.detailContent = fmt.Sprintf("Installation failed: %v", err)
			} else {
				// Re-detect
				status := core.DetectTool(*m.selected)
				m.statuses[m.selected.ID] = status
				m.buildDetail()
			}
		}
		m.currentScreen = screenDetail
	case "n", "N", "esc", "q":
		m.currentScreen = screenDetail
	}
	return m, nil
}

func (m *Model) handleRunKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.selected != nil {
			output, err := core.ExecuteTool(*m.selected, nil, core.ExecOptions{})
			if err != nil {
				m.detailContent = fmt.Sprintf("Execution error: %v\nOutput: %s", err, output)
			} else {
				m.detailContent = fmt.Sprintf("Output:\n%s", output)
			}
		}
		m.currentScreen = screenDetail
	case "n", "N", "esc", "q":
		m.currentScreen = screenDetail
	}
	return m, nil
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filtered = core.SearchTools(m.tools, m.searchQuery)
		m.cursor = 0
		m.currentScreen = screenList
	case "esc":
		m.currentScreen = screenList
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
	}
	return m, nil
}

func (m *Model) handleTerminalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.currentScreen = screenList
	}
	return m, nil
}

func (m *Model) View() string {
	switch m.currentScreen {
	case screenList:
		return m.listView()
	case screenDetail:
		return m.detailView()
	case screenInstall:
		return m.installView()
	case screenRun:
		return m.runView()
	case screenSearch:
		return m.searchView()
	case screenTerminal:
		return m.terminalView()
	case screenHelp:
		return m.helpView()
	}
	return ""
}

func (m *Model) listView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" AutoMate — Tool Manager "))
	b.WriteString("\n\n")

	// Stats bar
	installed := 0
	for _, s := range m.statuses {
		if s.OnPath || s.DockerImage || s.PackageManager != "" {
			installed++
		}
	}
	b.WriteString(statusStyle.Render(fmt.Sprintf("  %d tools • %d installed • Domains: %s",
		len(m.tools), installed, strings.Join(core.Domains(m.tools), ", "))))
	b.WriteString("\n\n")

	// Tool list
	for i, t := range m.filtered {
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
		}

		toolLine := prefix + t.Name
		if i == m.cursor {
			toolLine = selectedStyle.Render(prefix + t.Name)
		}

		// Status indicator
		statusStr := " "
		if s, ok := m.statuses[t.ID]; ok {
			if s.OnPath {
				statusStr = installedStyle.Render("✓")
			} else if !s.OnPath {
				statusStr = notInstalledStyle.Render("✗")
			}
		}

		desc := t.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}

		b.WriteString(fmt.Sprintf("  %s %s  %s\n", statusStr, toolLine, statusStyle.Render(desc)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  ↑↓ navigate • enter: details • /: search • i: install • r: run • t: terminal • ?: help • q: quit"))

	return b.String()
}

func (m *Model) buildDetail() {
	if m.selected == nil {
		return
	}

	t := m.selected
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n  %s (%s)\n\n", t.Name, t.ID))
	b.WriteString(fmt.Sprintf("  Description: %s\n", t.Description))
	b.WriteString(fmt.Sprintf("  Namespace: %s\n", t.Namespace))
	b.WriteString(fmt.Sprintf("  Risk: %s  |  Trust: %s\n", t.RiskLevel, t.TrustLevel))

	// Capabilities
	if len(t.Capabilities) > 0 {
		b.WriteString(fmt.Sprintf("  Capabilities: %s\n", strings.Join(t.Capabilities, ", ")))
	}

	// Install status
	if s, ok := m.statuses[t.ID]; ok {
		if s.OnPath {
			b.WriteString(installedStyle.Render(fmt.Sprintf("\n  ✓ Installed at: %s\n", s.PathLocation)))
		} else {
			b.WriteString(notInstalledStyle.Render("\n  ✗ Not installed\n"))
		}
	}

	// Template
	if t.Execution.Template != "" {
		b.WriteString(fmt.Sprintf("\n  Template: %s\n", t.Execution.Template))
	}

	// Parameters
	if len(t.Parameters) > 0 {
		b.WriteString(fmt.Sprintf("\n  Parameters (%d):\n", len(t.Parameters)))
		for _, p := range t.Parameters {
			req := " "
			if p.Required {
				req = "*"
			}
			def := ""
			if p.DefaultValue != nil {
				def = fmt.Sprintf(" [default: %v]", p.DefaultValue)
			}
			b.WriteString(fmt.Sprintf("    %s %s (%s)%s — %s\n", req, p.Name, p.Type, def, p.Description))
		}
	}

	// Install methods
	if len(t.Install) > 0 {
		b.WriteString(fmt.Sprintf("\n  Install methods (%d):\n", len(t.Install)))
		for _, inst := range t.Install {
			b.WriteString(fmt.Sprintf("    • %s", inst.Method))
			if inst.PackageName != "" {
				b.WriteString(fmt.Sprintf(" (%s)", inst.PackageName))
			}
			b.WriteString("\n")
		}
	}

	m.detailContent = b.String()
}

func (m *Model) detailView() string {
	var b strings.Builder
	b.WriteString(detailTitleStyle.Render(" Tool Details "))
	b.WriteString("\n")
	b.WriteString(m.detailContent)
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  i: install  •  r: run  •  enter: recheck  •  esc/q: back"))

	return b.String()
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
		b.WriteString(fmt.Sprintf("    %s", inst.Method))
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

	return b.String()
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
		return b.String()
	}

	b.WriteString(fmt.Sprintf("  Template: %s\n\n", m.selected.Execution.Template))

	// Show params
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

	return b.String()
}

func (m *Model) searchView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Search "))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  Query: %s█\n\n", m.searchQuery))
	b.WriteString(helpStyle.Render("  Type to search • enter: search • esc: cancel"))

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

	return b.String()
}

func (m *Model) helpView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Help "))
	b.WriteString("\n\n")
	b.WriteString("  Navigation:\n")
	b.WriteString("    ↑/↓ or j/k    — Move cursor\n")
	b.WriteString("    enter           — Select / open details\n")
	b.WriteString("    /               — Search tools\n")
	b.WriteString("    q / ctrl+c      — Quit\n\n")
	b.WriteString("  Actions:\n")
	b.WriteString("    i               — Install selected tool\n")
	b.WriteString("    r               — Run selected tool\n")
	b.WriteString("    t               — Open interactive terminal\n")
	b.WriteString("    R               — Refresh installation status\n\n")
	b.WriteString("  Screens:\n")
	b.WriteString("    esc / backspace — Go back\n")
	b.WriteString("    ?               — Toggle this help\n\n")
	b.WriteString("  Press any key to close help.\n")

	return b.String()
}

func main() {
	fmt.Println("This package should be imported, not run directly.")
	os.Exit(1)
}

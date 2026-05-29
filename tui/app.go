package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Intro-Shlok/AutoMate/core"
)

const (
	bootDuration = 2500 * time.Millisecond
	appVersion   = "0.1.0"
)

type screen int

const (
	screenBoot screen = iota
	screenList
	screenDetail
	screenInstall
	screenInstallProgress
	screenRun
	screenSearch
	screenTerminal
	screenHelp
)

type bootDoneMsg struct{}
type leaderTimeoutMsg struct{}
type spinnerTickMsg struct{}

type installProgressMsg struct {
	output string
	err    error
	done   bool
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type sysInfo struct {
	hostname  string
	os        string
	kernel    string
	arch      string
	cpuCores  int
	memory    string
	uptime    string
	toolCount int
	buildVer  string
	timeStr   string
}

type Model struct {
	tools      []core.ToolDefinition
	toolsCache *core.Cache
	cursor     int
	filtered   []core.ToolDefinition
	selected   *core.ToolDefinition
	searchQuery string
	statuses   map[string]core.InstallStatus
	detailContent string
	width      int
	height     int
	sysInfo    sysInfo

	currentScreen    screen
	leaderActive     bool
	sidebarVisible   bool
	selectedDomain   string

	paletteVisible  bool
	paletteQuery    string
	paletteResults  []core.ToolDefinition
	paletteCursor   int

	installProgress bytes.Buffer
	installOutput   string
	installingTool  *core.ToolDefinition
	installMethod   string
	spinnerIndex    int
}

func NewApp(tools []core.ToolDefinition, cache *core.Cache) *Model {
	return &Model{
		tools:         tools,
		toolsCache:    cache,
		filtered:      tools,
		statuses:      core.DetectAllTools(tools),
		currentScreen: screenBoot,
		sysInfo:       collectSysInfo(tools),
	}
}

func (m *Model) Run() error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m *Model) Init() tea.Cmd {
	return tea.Tick(bootDuration, func(t time.Time) tea.Msg {
		return bootDoneMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bootDoneMsg:
		if m.currentScreen == screenBoot {
			m.currentScreen = screenList
		}
		return m, nil

	case leaderTimeoutMsg:
		m.leaderActive = false
		return m, nil

	case spinnerTickMsg:
		if m.currentScreen == screenInstallProgress {
			m.spinnerIndex = (m.spinnerIndex + 1) % len(spinnerFrames)
			return m, spinnerTick()
		}
		return m, nil

	case installProgressMsg:
		if msg.done {
			m.currentScreen = screenDetail
			m.installOutput = msg.output
			if m.selected != nil {
				status := core.DetectTool(*m.selected)
				m.statuses[m.selected.ID] = status
				m.buildDetail()
			}
			return m, nil
		}
		if msg.output != "" {
			m.installOutput = msg.output
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Boot screen: any key skips
	if m.currentScreen == screenBoot {
		m.currentScreen = screenList
		return m, nil
	}

	// Leader key handling
	if msg.Type == tea.KeyCtrlX {
		m.leaderActive = true
		return m, tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
			return leaderTimeoutMsg{}
		})
	}

	// If leader is active, handle chord
	if m.leaderActive {
		m.leaderActive = false
		switch msg.String() {
		case "b":
			m.sidebarVisible = !m.sidebarVisible
			return m, nil
		case "s":
			m.currentScreen = screenHelp
			return m, nil
		case "p":
			m.openPalette()
			return m, nil
		case "n":
			m.currentScreen = screenList
			m.selected = nil
			return m, nil
		default:
			return m, nil
		}
	}

	// Command palette mode
	if m.paletteVisible {
		return m.handlePaletteKey(msg)
	}

	// Screen-specific handling
	switch m.currentScreen {
	case screenList:
		return m.handleListKey(msg)
	case screenDetail:
		return m.handleDetailKey(msg)
	case screenInstall:
		return m.handleInstallKey(msg)
	case screenInstallProgress:
		if msg.String() == "esc" {
			m.currentScreen = screenDetail
		}
		return m, nil
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
		m.statuses = core.DetectAllTools(m.tools)
	case "ctrl+p":
		m.openPalette()
		return m, nil
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
			m.buildDetail()
		}
	case "ctrl+p":
		m.openPalette()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleInstallKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.selected != nil {
			m.installingTool = m.selected
			m.installOutput = ""
			m.installProgress.Reset()
			m.currentScreen = screenInstallProgress
			return m, m.runInstall(*m.selected)
		}
		m.currentScreen = screenDetail
	case "n", "N", "esc", "q":
		m.currentScreen = screenDetail
	}
	return m, nil
}

func spinnerTick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m *Model) runInstall(tool core.ToolDefinition) tea.Cmd {
	return tea.Batch(
		spinnerTick(),
		func() tea.Msg {
			var buf bytes.Buffer
			opts := core.InstallOptions{
				Output: &buf,
			}
			_ = core.InstallTool(tool, opts)

			return installProgressMsg{
				output: buf.String(),
				done:   true,
			}
		},
	)
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

// Command palette handling

func (m *Model) openPalette() {
	m.paletteVisible = true
	m.paletteQuery = ""
	m.paletteResults = m.tools
	m.paletteCursor = 0
}

func (m *Model) closePalette() {
	m.paletteVisible = false
	m.paletteQuery = ""
	m.paletteResults = nil
	m.paletteCursor = 0
}

func (m *Model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.closePalette()
		return m, nil
	case "enter":
		if len(m.paletteResults) > 0 && m.paletteCursor < len(m.paletteResults) {
			m.selected = &m.paletteResults[m.paletteCursor]
			m.buildDetail()
			m.closePalette()
			m.currentScreen = screenDetail
		}
		return m, nil
	case "up", "k":
		if m.paletteCursor > 0 {
			m.paletteCursor--
		}
		return m, nil
	case "down", "j":
		if m.paletteCursor < len(m.paletteResults)-1 {
			m.paletteCursor++
		}
		return m, nil
	case "backspace":
		if len(m.paletteQuery) > 0 {
			m.paletteQuery = m.paletteQuery[:len(m.paletteQuery)-1]
			m.filterPalette()
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.paletteQuery += msg.String()
			m.filterPalette()
		}
		return m, nil
	}
}

func (m *Model) filterPalette() {
	m.paletteResults = core.SearchTools(m.tools, m.paletteQuery)
	m.paletteCursor = 0
}

// View method

func (m *Model) View() string {
	var content string

	if m.paletteVisible {
		content = m.baseView() + "\n\n" + m.paletteView()
	} else if m.currentScreen == screenDetail {
		// Split layout: compact list on left, detail on right
		listWidth := 40
		left := m.compactListView(listWidth)
		right := m.detailView()
		content = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	} else {
		content = m.baseView()
	}

	if m.currentScreen != screenBoot {
		content += "\n" + m.statusBarView()
	}

	return content
}

func (m *Model) baseView() string {
	switch m.currentScreen {
	case screenBoot:
		return m.bootView()
	case screenList:
		return m.listView()
	case screenDetail:
		return m.detailView()
	case screenInstall:
		return m.installView()
	case screenInstallProgress:
		return m.installProgressView()
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

func (m *Model) buildDetail() {
	if m.selected == nil {
		return
	}
	// Detail is built by the view function now
}

func (m *Model) setDomainFilter(domain string) {
	m.selectedDomain = domain
	m.filtered = toolsForDomain(m.tools, domain)
	m.cursor = 0
}

func collectSysInfo(tools []core.ToolDefinition) sysInfo {
	info := sysInfo{
		arch:      runtime.GOARCH,
		cpuCores:  runtime.NumCPU(),
		toolCount: len(tools),
		buildVer:  appVersion,
		timeStr:   time.Now().Local().Format("Mon Jan 2 15:04:05 MST 2006"),
	}

	host, err := os.Hostname()
	if err == nil {
		info.hostname = host
	} else {
		info.hostname = "unknown"
	}

	switch runtime.GOOS {
	case "linux":
		info.os = "Linux"

		kernel, _ := os.ReadFile("/proc/sys/kernel/ostype")
		ver, _ := os.ReadFile("/proc/sys/kernel/osrelease")
		if len(kernel) > 0 && len(ver) > 0 {
			info.kernel = strings.TrimSpace(string(kernel)) + " " + strings.TrimSpace(string(ver))
		} else {
			if out, err := exec.Command("uname", "-sr").Output(); err == nil {
				info.kernel = strings.TrimSpace(string(out))
			}
		}

		mem, _ := os.ReadFile("/proc/meminfo")
		if len(mem) > 0 {
			for _, line := range strings.Split(string(mem), "\n") {
				if strings.HasPrefix(line, "MemTotal:") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						if kb, err := strconv.Atoi(parts[1]); err == nil {
							mb := kb / 1024
							if mb > 1024 {
								info.memory = fmt.Sprintf("%.1f GB", float64(mb)/1024)
							} else {
								info.memory = fmt.Sprintf("%d MB", mb)
							}
						}
					}
					break
				}
			}
		}

		uptimeBytes, _ := os.ReadFile("/proc/uptime")
		if len(uptimeBytes) > 0 {
			parts := strings.Fields(string(uptimeBytes))
			if len(parts) > 0 {
				if secs, err := strconv.ParseFloat(parts[0], 64); err == nil {
					d := time.Duration(secs) * time.Second
					days := int(d.Hours()) / 24
					hrs := int(d.Hours()) % 24
					mins := int(d.Minutes()) % 60
					if days > 0 {
						info.uptime = fmt.Sprintf("%dd %dh %dm", days, hrs, mins)
					} else if hrs > 0 {
						info.uptime = fmt.Sprintf("%dh %dm", hrs, mins)
					} else {
						info.uptime = fmt.Sprintf("%dm", mins)
					}
				}
			}
		}

	case "darwin":
		info.os = "macOS"
		if out, err := exec.Command("uname", "-sr").Output(); err == nil {
			info.kernel = strings.TrimSpace(string(out))
		}
		if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
			if b, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
				gb := float64(b) / 1e9
				info.memory = fmt.Sprintf("%.1f GB", gb)
			}
		}
		if out, err := exec.Command("sysctl", "-n", "kern.boottime").Output(); err == nil {
			parts := strings.Fields(string(out))
			for i, p := range parts {
				if p == "sec" && i+2 < len(parts) {
					if secs, err := strconv.ParseInt(strings.TrimRight(parts[i+1], ","), 10, 64); err == nil {
						boot := time.Unix(secs, 0)
						d := time.Since(boot)
						days := int(d.Hours()) / 24
						hrs := int(d.Hours()) % 24
						mins := int(d.Minutes()) % 60
						if days > 0 {
							info.uptime = fmt.Sprintf("%dd %dh %dm", days, hrs, mins)
						} else if hrs > 0 {
							info.uptime = fmt.Sprintf("%dh %dm", hrs, mins)
						} else {
							info.uptime = fmt.Sprintf("%dm", mins)
						}
					}
					break
				}
			}
		}

	default:
		info.os = runtime.GOOS
		if out, err := exec.Command("uname", "-sr").Output(); err == nil {
			info.kernel = strings.TrimSpace(string(out))
		}
	}

	if info.kernel == "" {
		info.kernel = runtime.GOOS
	}
	if info.memory == "" {
		info.memory = "unknown"
	}
	if info.uptime == "" {
		info.uptime = "unknown"
	}

	return info
}

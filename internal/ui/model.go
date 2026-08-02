package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
)

const tickInterval = 500 * time.Millisecond

type tickMsg time.Time

var (
	accent       = lipgloss.Color("212")
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	tabStyle     = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("240"))
	activeTab    = lipgloss.NewStyle().Padding(0, 1).Background(accent).Foreground(lipgloss.Color("0")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selStyle     = lipgloss.NewStyle().Foreground(accent).Bold(true)
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	spinnerStyle = lipgloss.NewStyle().Foreground(accent)
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	boxStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 2)
)

var tabNames = []string{"Home", "Store", "Extensions", "Downloads", "Settings"}

type Model struct {
	cfg  app.Config
	qm   *queue.Manager
	w, h int
	tab  int
	msg  string

	// geometry recorded at render time for mouse hit-testing
	boxTop, boxLeft int
	tabXs           [][2]int
	listY           int // content line where the selectable list starts
	listWinStart    int // absolute index of the first rendered list row

	home     homeState
	store    storeState
	exts     extState
	dl       dlState
	settings settingsState
	auth     *authState
}

type homeState struct {
	input        textinput.Model
	inputFocused bool
	results      []backend.Track
	cursor       int
	kind         string
	loading      bool
	spin         spinner.Model
}

type storeState struct {
	input        textinput.Model
	inputFocused bool
	exts         []backend.RepoExt
	cursor       int
	loading      bool
	spin         spinner.Model
}

type extState struct {
	exts     []backend.InstalledExt
	cursor   int
	priority []string
}

type dlState struct {
	cursor int
}

type settingsState struct {
	input        textinput.Model
	inputFocused bool
}

type authState struct {
	req   backend.AuthRequest
	input textinput.Model
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return s
}

func newInput(placeholder string) textinput.Model {
	i := textinput.New()
	i.Placeholder = placeholder
	i.Width = 60
	i.CharLimit = 512
	return i
}

func New(cfg app.Config, qm *queue.Manager) *Model {
	m := &Model{cfg: cfg, qm: qm, listY: -1}
	m.home.input = newInput("URL or search query")
	m.home.spin = newSpinner()
	m.store.input = newInput("https://github.com/owner/repo  (extension registry)")
	m.store.spin = newSpinner()
	m.settings.input = newInput("output directory")
	m.settings.input.SetValue(cfg.OutputDir)
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) }),
		loadStore(), loadExts(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		w := clamp(msg.Width-10, 20, 100)
		m.home.input.Width = w
		m.store.input.Width = w
		m.settings.input.Width = w
		if m.auth != nil {
			m.auth.input.Width = w
		}
		return m, nil

	case tickMsg:
		m.qm.Refresh()
		return m, tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })

	case spinner.TickMsg:
		var cmds []tea.Cmd
		if m.home.loading {
			m.home.spin, _ = m.home.spin.Update(msg)
			cmds = append(cmds, func() tea.Msg { return m.home.spin.Tick() })
		}
		if m.store.loading {
			m.store.spin, _ = m.store.spin.Update(msg)
			cmds = append(cmds, func() tea.Msg { return m.store.spin.Tick() })
		}
		return m, tea.Batch(cmds...)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		if m.auth != nil {
			return m.updateAuth(msg)
		}
		if inp := m.focusedInput(); inp != nil {
			return m.updateFocusedInput(msg)
		}
		return m.updateGlobal(msg)
	}

	return m.handleAsync(msg)
}

func (m *Model) focusedInput() *textinput.Model {
	switch {
	case m.tab == 0 && m.home.inputFocused:
		return &m.home.input
	case m.tab == 1 && m.store.inputFocused:
		return &m.store.input
	case m.tab == 4 && m.settings.inputFocused:
		return &m.settings.input
	}
	return nil
}

func (m *Model) updateFocusedInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.blurInput()
		return m, nil
	case tea.KeyEnter:
		return m.submitFocusedInput()
	}
	var cmd tea.Cmd
	if inp := m.focusedInput(); inp != nil {
		*inp, cmd = inp.Update(msg)
	}
	return m, cmd
}

func (m *Model) blurInput() {
	switch m.tab {
	case 0:
		m.home.input.Blur()
		m.home.inputFocused = false
	case 1:
		m.store.input.Blur()
		m.store.inputFocused = false
	case 4:
		m.settings.input.Blur()
		m.settings.inputFocused = false
		m.settings.input.SetValue(m.cfg.OutputDir)
	}
}

func (m *Model) focusInput() {
	switch m.tab {
	case 0:
		m.home.input.Focus()
		m.home.inputFocused = true
	case 1:
		m.store.input.Focus()
		m.store.inputFocused = true
	case 4:
		m.settings.input.Focus()
		m.settings.inputFocused = true
	}
}

func (m *Model) submitFocusedInput() (tea.Model, tea.Cmd) {
	switch m.tab {
	case 0:
		query := strings.TrimSpace(m.home.input.Value())
		m.home.input.SetValue("")
		m.home.inputFocused = false
		m.home.loading = true
		return m, tea.Batch(resolveCmd(query), func() tea.Msg { return m.home.spin.Tick() })
	case 1:
		url := strings.TrimSpace(m.store.input.Value())
		m.store.input.SetValue("")
		m.store.inputFocused = false
		if url == "" {
			return m, nil
		}
		if err := backend.AddRegistry(url); err != nil {
			m.msg = "registry: " + err.Error()
			return m, nil
		}
		m.store.loading = true
		return m, tea.Batch(func() tea.Msg { return m.store.spin.Tick() }, loadStore())
	case 4:
		dir := strings.TrimSpace(m.settings.input.Value())
		m.cfg.OutputDir = dir
		m.settings.inputFocused = false
		return m, m.saveCmd()
	}
	return m, nil
}

func (m *Model) updateGlobal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlD:
		return m, tea.Quit
	case tea.KeyTab:
		m.tab = (m.tab + 1) % 5
		return m, nil
	case tea.KeyShiftTab:
		m.tab = (m.tab + 4) % 5
		return m, nil
	case tea.KeyCtrlL:
		if err := m.cfg.Save(); err != nil {
			m.msg = err.Error()
		} else {
			m.msg = "settings saved"
		}
		return m, nil
	case tea.KeyRunes:
		switch msg.String() {
		case "1", "2", "3", "4", "5":
			m.tab = int(msg.String()[0] - '1')
			return m, nil
		case "q":
			return m, tea.Quit
		case "?":
			m.msg = "tab/1-5 switch · enter submit · esc cancel · q quit · mouse works too"
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.tab {
	case 0:
		cmd = m.homeAction(msg)
	case 1:
		cmd = m.storeAction(msg)
	case 2:
		cmd = m.extsAction(msg)
	case 3:
		cmd = m.dlAction(msg)
	case 4:
		cmd = m.settingsAction(msg)
	}
	return m, cmd
}

func (m *Model) handleAsync(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case storeLoadedMsg:
		m.store.exts = msg.exts
		m.store.loading = false
		if msg.err != nil {
			m.msg = "store: " + msg.err.Error()
		}
	case extsLoadedMsg:
		m.exts.exts = msg.exts
		m.exts.priority = msg.priority
		if msg.err != nil {
			m.msg = "extensions: " + msg.err.Error()
		}
	case resultsMsg:
		m.home.loading = false
		m.home.results = msg.tracks
		m.home.cursor = 0
		m.home.kind = msg.kind
		if msg.err != nil {
			m.home.results = nil
			m.msg = msg.err.Error()
		}
	case enqueuedMsg:
		m.msg = fmt.Sprintf("enqueued %d", msg.n)
	case savedMsg:
		m.msg = "settings saved"
	case authPendingMsg:
		if msg.err != nil {
			m.msg = "auth: " + msg.err.Error()
			return m, nil
		}
		if msg.req.AuthURL == "" {
			m.msg = "no pending auth requests"
			return m, nil
		}
		m.auth = &authState{req: msg.req, input: newInput("paste auth code or callback URL")}
		m.auth.input.Focus()
	}
	return m, nil
}

// ---- View + centered modern layout ----

func (m *Model) View() string {
	if m.w == 0 || m.h == 0 {
		return "loading…"
	}
	m.listY = -1
	content := m.renderHeader() + "\n\n" + m.content() + "\n\n" + m.renderFooter()
	maxW := clamp(m.w-6, 40, 110)
	boxed := boxStyle.Width(maxW).Render(content)
	boxH := lipgloss.Height(boxed)
	boxW := lipgloss.Width(boxed)
	m.boxTop = max((m.h-boxH)/2, 0)
	m.boxLeft = max((m.w-boxW)/2, 0)
	m.tabXs = m.computeTabXs()
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, boxed)
}

func (m *Model) content() string {
	if m.auth != nil {
		return m.renderAuth()
	}
	switch m.tab {
	case 0:
		return m.renderHome()
	case 1:
		return m.renderStore()
	case 2:
		return m.renderExts()
	case 3:
		return m.renderDownloads()
	case 4:
		return m.renderSettings()
	}
	return ""
}

func (m *Model) renderHeader() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" SpotiFLAC TUI "))
	for i, t := range tabNames {
		if i == m.tab {
			b.WriteString(activeTab.Render(fmt.Sprintf(" %d %s ", i+1, t)))
		} else {
			b.WriteString(tabStyle.Render(fmt.Sprintf(" %d %s ", i+1, t)))
		}
	}
	return b.String()
}

func (m *Model) computeTabXs() [][2]int {
	title := titleStyle.Render(" SpotiFLAC TUI ")
	x := lipgloss.Width(title)
	var out [][2]int
	for i, t := range tabNames {
		var piece string
		if i == m.tab {
			piece = activeTab.Render(fmt.Sprintf(" %d %s ", i+1, t))
		} else {
			piece = tabStyle.Render(fmt.Sprintf(" %d %s ", i+1, t))
		}
		w := lipgloss.Width(piece)
		out = append(out, [2]int{x, x + w})
		x += w
	}
	return out
}

func (m *Model) renderFooter() string {
	hints := map[int]string{
		0: "e search · j/k or wheel · d download · a download all",
		1: "r add repo · i install · R refresh · s switch repo",
		2: "x toggle · j/k move prio · m save · v verify",
		3: "j/k or wheel · c cancel",
		4: "e edit dir · q quality · m metadata · l lyrics · s save",
	}
	h := hints[m.tab]
	if m.msg != "" {
		h = m.msg + "   ·   " + h
	}
	return helpStyle.Render(h)
}

// ---- Mouse ----

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	contentX := msg.X - m.boxLeft - 3 // border(1) + left padding(2)
	contentY := msg.Y - m.boxTop - 1  // border(1)

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if m.tab == 0 {
			if m.home.cursor > 0 {
				m.home.cursor--
			}
		} else if m.tab == 1 {
			if m.store.cursor > 0 {
				m.store.cursor--
			}
		} else if m.tab == 2 {
			if m.exts.cursor > 0 {
				m.movePriority(-1)
				m.exts.cursor--
			}
		} else if m.tab == 3 {
			if m.dl.cursor > 0 {
				m.dl.cursor--
			}
		}
		return m, nil
	case tea.MouseButtonWheelDown:
		if m.tab == 0 {
			if m.home.cursor < len(m.home.results)-1 {
				m.home.cursor++
			}
		} else if m.tab == 1 {
			if m.store.cursor < len(m.store.exts)-1 {
				m.store.cursor++
			}
		} else if m.tab == 2 {
			if m.exts.cursor < len(m.exts.exts)-1 {
				m.movePriority(1)
				m.exts.cursor++
			}
		} else if m.tab == 3 {
			if m.dl.cursor < len(m.qm.Items())-1 {
				m.dl.cursor++
			}
		}
		return m, nil
	}

	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return m, nil
	}

	// header row: switch tab by x
	if contentY == 0 {
		for i, r := range m.tabXs {
			if contentX >= r[0] && contentX < r[1] {
				m.tab = i
				return m, nil
			}
		}
		return m, nil
	}

	// input row is the first line of tab content (contentY 2: header at 0, blank at 1)
	if contentY == 2 {
		if m.tab == 0 || m.tab == 1 || m.tab == 4 {
			m.focusInput()
			return m, nil
		}
	}

	// list rows (content starts at contentY 2: header + blank)
	if m.listY >= 0 {
		idx := contentY - 2 - m.listY
		if idx >= 0 {
			n := m.listLen()
			if m.listWinStart+idx < n {
				m.setCursor(m.listWinStart + idx)
			}
		}
	}
	return m, nil
}

func (m *Model) listLen() int {
	switch m.tab {
	case 0:
		return len(m.home.results)
	case 1:
		return len(m.store.exts)
	case 2:
		return len(m.exts.exts)
	case 3:
		return len(m.qm.Items())
	}
	return 0
}

func (m *Model) setCursor(i int) {
	switch m.tab {
	case 0:
		if i >= 0 && i < len(m.home.results) {
			m.home.cursor = i
		}
	case 1:
		if i >= 0 && i < len(m.store.exts) {
			m.store.cursor = i
		}
	case 2:
		if i >= 0 && i < len(m.exts.exts) {
			m.exts.cursor = i
		}
	case 3:
		if i >= 0 && i < len(m.qm.Items()) {
			m.dl.cursor = i
		}
	}
}

// listWindow returns the absolute start/end indexes to render for a list so it
// stays inside the visible box.
func (m *Model) listWindow(n, cursor, visible int) (start, end int) {
	if n <= visible {
		return 0, n
	}
	start = cursor - visible/2
	if start < 0 {
		start = 0
	}
	if start+visible > n {
		start = n - visible
	}
	return start, start + visible
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

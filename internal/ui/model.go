package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
)

type tickMsg time.Time

const tickInterval = 500 * time.Millisecond

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	tabStyle     = lipgloss.NewStyle().Padding(0, 1)
	activeTab    = tabStyle.Background(lipgloss.Color("212")).Foreground(lipgloss.Color("0"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
)

type Model struct {
	cfg  app.Config
	qm   *queue.Manager
	w, h int
	tab  int
	msg  string

	home     homeState
	store    storeState
	exts     extState
	dl       dlState
	settings settingsState

	auth *authState
}

type homeState struct {
	input   string
	typing  bool
	results []backend.Track
	cursor  int
	kind    string // "URL" or "search"
}

type storeState struct {
	input   string
	typing  bool
	exts    []backend.RepoExt
	cursor  int
	loading bool
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
	input      string
	typing     bool
	editQuality bool
}

type authState struct {
	req  backend.AuthRequest
	code string
}

func New(cfg app.Config, qm *queue.Manager) *Model {
	m := &Model{cfg: cfg, qm: qm}
	m.settings.input = cfg.OutputDir
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
	case tickMsg:
		m.qm.Refresh()
		return m, tea.Tick(tickInterval, func(t time.Time) tea.Msg { return tickMsg(t) })
	case tea.KeyMsg:
		if m.auth != nil {
			return m.handleAuthKey(msg)
		}
		return m.handleKey(msg)
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
	}
	return m, nil
}

func (m *Model) View() string {
	if m.w == 0 {
		return "loading..."
	}
	if m.auth != nil {
		return m.renderHeader() + "\n\n" + m.renderAuth() + "\n\n" +
			dimStyle.Render("enter submit · esc cancel") + "\n" + m.renderFooter()
	}
	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	switch m.tab {
	case 0:
		b.WriteString(m.renderHome())
	case 1:
		b.WriteString(m.renderStore())
	case 2:
		b.WriteString(m.renderExts())
	case 3:
		b.WriteString(m.renderDownloads())
	case 4:
		b.WriteString(m.renderSettings())
	}
	b.WriteString(m.renderFooter())
	return b.String()
}

func (m *Model) renderHeader() string {
	tabs := []string{"Home", "Store", "Extensions", "Downloads", "Settings"}
	var b strings.Builder
	b.WriteString(titleStyle.Render(" SpotiFLAC TUI "))
	for i, t := range tabs {
		if i == m.tab {
			b.WriteString(activeTab.Render(fmt.Sprintf(" %d %s ", i+1, t)))
		} else {
			b.WriteString(tabStyle.Render(fmt.Sprintf(" %d %s ", i+1, t)))
		}
	}
	return b.String()
}

func (m *Model) renderFooter() string {
	hints := map[int]string{
		0: "[enter] go  [d] dl  [a] dl all  [e] type  [esc] stop typing",
		1: "[i] install  [r] add repo  [R] refresh  [s] switch repo",
		2: "[x] toggle  [up/dn] move prio  [m] save prio",
		3: "[c] cancel  [j/k] select",
		4: "[s] save  [q] quality:",
	}
	msg := m.msg
	if msg == "" {
		msg = hints[m.tab]
	}
	return "\n" + dimStyle.Render(msg) + "\n" + dimStyle.Render("ctrl+c quit  ·  1-5 tab")
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyCtrlL:
		if err := m.cfg.Save(); err != nil {
			m.msg = err.Error()
		}
		return m, nil
	case tea.KeyRunes:
		switch msg.String() {
		case "1", "2", "3", "4", "5":
			if !m.typing() {
				m.tab = int(msg.String()[0] - '1')
				return m, nil
			}
		case "q":
			if !m.typing() {
				return m, tea.Quit
			}
		}
	case tea.KeyTab:
		if !m.typing() {
			m.tab = (m.tab + 1) % 5
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.tab {
	case 0:
		cmd = m.homeKey(msg)
	case 1:
		cmd = m.storeKey(msg)
	case 2:
		cmd = m.extsKey(msg)
	case 3:
		cmd = m.dlKey(msg)
	case 4:
		cmd = m.settingsKey(msg)
	}
	return m, cmd
}

func (m *Model) typing() bool {
	switch m.tab {
	case 0:
		return m.home.typing
	case 1:
		return m.store.typing
	case 4:
		return m.settings.typing
	}
	return false
}

// typeInto appends characters to *s based on a key message. Returns true if handled.
func typeInto(s *string, msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyBackspace:
		if len(*s) > 0 {
			*s = (*s)[:len(*s)-1]
		}
		return true
	case tea.KeySpace:
		*s += " "
		return true
	case tea.KeyEnter, tea.KeyEsc:
		return false // leave to caller
	case tea.KeyRunes:
		*s += msg.String()
		return true
	}
	return false
}

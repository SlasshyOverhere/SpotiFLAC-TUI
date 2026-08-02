package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/backend"
)

type extsLoadedMsg struct {
	exts     []backend.InstalledExt
	priority []string
	err      error
}

func loadExts() tea.Cmd {
	return func() tea.Msg {
		exts, err := backend.ListInstalled()
		prio, _ := backend.ProviderPriority()
		return extsLoadedMsg{exts: exts, priority: prio, err: err}
	}
}

func (m *Model) renderExts() string {
	var b strings.Builder
	m.listY = 0
	if len(m.exts.priority) > 0 {
		b.WriteString(dimStyle.Render("provider priority: "))
		b.WriteString(strings.Join(m.exts.priority, " > "))
		b.WriteString("\n\n")
		m.listY = 2
	}
	if len(m.exts.exts) == 0 {
		b.WriteString(dimStyle.Render("no extensions installed — install some from the Store tab"))
		return b.String()
	}
	visible := clamp(m.h-12, 5, 30)
	start, end := m.listWindow(len(m.exts.exts), m.exts.cursor, visible)
	m.listWinStart = start
	for i := start; i < end; i++ {
		e := m.exts.exts[i]
		line := fmt.Sprintf("%s  v%s", e.DisplayName, e.Version)
		types := []string{}
		if e.HasMetadataProvider {
			types = append(types, "meta")
		}
		if e.HasDownloadProvider {
			types = append(types, "dl")
		}
		if e.HasLyricsProvider {
			types = append(types, "lyrics")
		}
		if len(types) > 0 {
			line += dimStyle.Render(" [" + strings.Join(types, ",") + "]")
		}
		state := dimStyle.Render("off")
		if e.Enabled {
			state = okStyle.Render("on")
		}
		if e.Status == "error" {
			line += errStyle.Render("  error: " + e.Error)
		}
		if i == m.exts.cursor {
			b.WriteString(selStyle.Render("▸ "))
			b.WriteString(line + "  " + state)
		} else {
			b.WriteString("  " + line + "  " + state)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Model) extsAction(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyUp:
		if m.exts.cursor > 0 {
			m.movePriority(-1)
			m.exts.cursor--
		}
	case tea.KeyDown:
		if m.exts.cursor < len(m.exts.exts)-1 {
			m.movePriority(1)
			m.exts.cursor++
		}
	case tea.KeyRunes:
		switch msg.String() {
		case "x":
			if len(m.exts.exts) > 0 {
				e := m.exts.exts[m.exts.cursor]
				if err := backend.SetEnabled(e.ID, !e.Enabled); err != nil {
					m.msg = err.Error()
				}
				return loadExts()
			}
		case "m":
			if err := backend.SetProviderPriority(m.exts.priority); err != nil {
				m.msg = err.Error()
			} else {
				m.msg = "priority saved"
			}
		case "v":
			return m.authCmd()
		case "k":
			if m.exts.cursor > 0 {
				m.movePriority(-1)
				m.exts.cursor--
			}
		case "j":
			if m.exts.cursor < len(m.exts.exts)-1 {
				m.movePriority(1)
				m.exts.cursor++
			}
		}
	}
	return nil
}

func (m *Model) movePriority(dir int) {
	if len(m.exts.exts) == 0 {
		return
	}
	cur := m.exts.exts[m.exts.cursor].ID
	idx := indexOf(m.exts.priority, cur)
	if idx < 0 {
		m.exts.priority = append(m.exts.priority, cur)
		idx = len(m.exts.priority) - 1
	}
	target := idx + dir
	if target < 0 || target >= len(m.exts.priority) {
		return
	}
	m.exts.priority[idx], m.exts.priority[target] = m.exts.priority[target], m.exts.priority[idx]
}

func indexOf(xs []string, v string) int {
	for i, x := range xs {
		if x == v {
			return i
		}
	}
	return -1
}

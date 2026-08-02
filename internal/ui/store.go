package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/backend"
)

type storeLoadedMsg struct {
	exts []backend.RepoExt
	err  error
}

func loadStore() tea.Cmd {
	return func() tea.Msg {
		exts, err := backend.ListRepoExtensions(false)
		return storeLoadedMsg{exts: exts, err: err}
	}
}

func (m *Model) renderStore() string {
	var b strings.Builder
	if m.store.inputFocused {
		b.WriteString(selStyle.Render("▸ ") + m.store.input.View())
	} else {
		b.WriteString(dimStyle.Render("  " + m.store.input.View()))
	}
	b.WriteString("\n")

	regs, _ := backend.ListRegistries()
	m.listY = 2
	if len(regs) > 0 {
		b.WriteString(dimStyle.Render("  repo: " + regs[len(regs)-1]))
		b.WriteString("\n\n")
		m.listY = 3
	} else {
		b.WriteString("\n")
	}

	if m.store.loading {
		b.WriteString(m.store.spin.View() + " loading…")
		return b.String()
	}
	if len(m.store.exts) == 0 {
		b.WriteString(dimStyle.Render("no extensions — press r to add a registry URL"))
		return b.String()
	}
	visible := m.h - 14
	start, end := m.listWindow(len(m.store.exts), m.store.cursor, visible)
	m.listWinStart = start
	for i := start; i < end; i++ {
		e := m.store.exts[i]
		line := fmt.Sprintf("%s  v%s", e.DisplayName, e.Version)
		if e.Name != "" && e.Name != e.DisplayName {
			line += fmt.Sprintf("  [%s]", e.Name)
		}
		switch {
		case e.IsInstalled && e.HasUpdate:
			line += okStyle.Render("  UPDATE")
		case e.IsInstalled:
			line += dimStyle.Render("  installed")
		}
		if e.Downloads > 0 {
			line += dimStyle.Render(fmt.Sprintf("  ⤓%d", e.Downloads))
		}
		if i == m.store.cursor {
			b.WriteString(selStyle.Render("▸ " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Model) storeAction(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyUp:
		if m.store.cursor > 0 {
			m.store.cursor--
		}
	case tea.KeyDown:
		if m.store.cursor < len(m.store.exts)-1 {
			m.store.cursor++
		}
	case tea.KeyRunes:
		switch msg.String() {
		case "r":
			m.store.input.Focus()
			m.store.inputFocused = true
		case "R":
			m.store.loading = true
			return tea.Batch(func() tea.Msg { return m.store.spin.Tick() }, loadStore())
		case "i":
			if len(m.store.exts) > 0 {
				id := m.store.exts[m.store.cursor].ID
				if err := backend.InstallRepoExt(id); err != nil {
					m.msg = "install " + id + ": " + err.Error()
				} else {
					m.msg = "installed " + id
				}
				m.store.loading = true
				return tea.Batch(func() tea.Msg { return m.store.spin.Tick() }, loadStore())
			}
		case "s":
			regs, err := backend.ListRegistries()
			if err == nil && len(regs) > 1 {
				last := regs[len(regs)-1]
				for i, r := range regs {
					if r == last {
						next := (i + 1) % len(regs)
						if err := backend.SetRegistry(next); err == nil {
							m.msg = "repo: " + regs[next]
						}
						break
					}
				}
				m.store.loading = true
				return tea.Batch(func() tea.Msg { return m.store.spin.Tick() }, loadStore())
			}
		case "k":
			if m.store.cursor > 0 {
				m.store.cursor--
			}
		case "j":
			if m.store.cursor < len(m.store.exts)-1 {
				m.store.cursor++
			}
		}
	}
	return nil
}

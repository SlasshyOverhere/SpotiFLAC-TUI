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
	regs, _ := backend.ListRegistries()
	active := ""
	if len(regs) > 0 {
		active = regs[len(regs)-1]
	}
	if m.store.typing {
		b.WriteString(dimStyle.Render("Repository URL: "))
		b.WriteString(selStyle.Render(m.store.input + "▏"))
	} else {
		b.WriteString(dimStyle.Render("repo: "))
		b.WriteString(active)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if m.store.loading {
		b.WriteString(dimStyle.Render("loading..."))
		return b.String()
	}
	for i, e := range m.store.exts {
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

func (m *Model) storeKey(msg tea.KeyMsg) tea.Cmd {
	if m.store.typing {
		if msg.Type == tea.KeyEsc {
			m.store.typing = false
			return nil
		}
		if msg.Type == tea.KeyEnter {
			url := strings.TrimSpace(m.store.input)
			m.store.typing = false
			m.store.input = ""
			if url != "" {
				if err := backend.AddRegistry(url); err != nil {
					m.msg = err.Error()
					return nil
				}
				m.store.loading = true
				return loadStore()
			}
			return nil
		}
		if typeInto(&m.store.input, msg) {
			return nil
		}
		return nil
	}
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
			m.store.typing = true
		case "R":
			m.store.loading = true
			return loadStore()
		case "i":
			if len(m.store.exts) > 0 {
				id := m.store.exts[m.store.cursor].ID
				if err := backend.InstallRepoExt(id); err != nil {
					m.msg = "install " + id + ": " + err.Error()
				} else {
					m.msg = "installed " + id
				}
				m.store.loading = true
				return loadStore()
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
				return loadStore()
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

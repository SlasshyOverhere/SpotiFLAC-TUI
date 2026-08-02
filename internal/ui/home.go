package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
)

type resultsMsg struct {
	tracks []backend.Track
	kind   string
	err    error
}

type enqueuedMsg struct{ n int }

func (m *Model) renderHome() string {
	var b strings.Builder
	if m.home.inputFocused {
		b.WriteString(selStyle.Render("▸ ") + m.home.input.View())
	} else {
		b.WriteString(dimStyle.Render("  " + m.home.input.View()))
	}
	b.WriteString("\n\n")

	m.listY = 2
	if m.home.loading {
		b.WriteString(m.home.spin.View() + " working…")
		return b.String()
	}
	if m.home.kind != "" {
		b.WriteString(dimStyle.Render(m.home.kind))
		b.WriteString("\n")
		m.listY = 3
	}
	if len(m.home.results) == 0 {
		b.WriteString(dimStyle.Render("no results yet — type a URL or query and press enter"))
		return b.String()
	}
	visible := m.h - 12
	start, end := m.listWindow(len(m.home.results), m.home.cursor, visible)
	m.listWinStart = start
	for i := start; i < end; i++ {
		t := m.home.results[i]
		line := fmt.Sprintf("%s — %s", t.Name, t.Artists)
		if t.AlbumName != "" {
			line += fmt.Sprintf("  (%s)", t.AlbumName)
		}
		if i == m.home.cursor {
			b.WriteString(selStyle.Render("▸ " + line))
		} else {
			b.WriteString("  " + line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m *Model) homeAction(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEnter:
		return m.downloadSelected()
	case tea.KeyUp:
		if m.home.cursor > 0 {
			m.home.cursor--
		}
	case tea.KeyDown:
		if m.home.cursor < len(m.home.results)-1 {
			m.home.cursor++
		}
	case tea.KeyRunes:
		switch msg.String() {
		case "e":
			m.home.input.Focus()
			m.home.inputFocused = true
		case "d":
			return m.downloadSelected()
		case "a":
			if len(m.home.results) > 0 {
				for _, t := range m.home.results {
					req := backend.DownloadRequestForTrack(t, m.cfg)
					m.qm.Enqueue(queue.Queued{Title: t.Name, RequestJSON: backend.DownloadRequestJSON(req)})
				}
				return func() tea.Msg { return enqueuedMsg{len(m.home.results)} }
			}
		case "k":
			if m.home.cursor > 0 {
				m.home.cursor--
			}
		case "j":
			if m.home.cursor < len(m.home.results)-1 {
				m.home.cursor++
			}
		}
	}
	return nil
}

func (m *Model) downloadSelected() tea.Cmd {
	if len(m.home.results) == 0 || m.home.cursor < 0 || m.home.cursor >= len(m.home.results) {
		return nil
	}
	t := m.home.results[m.home.cursor]
	req := backend.DownloadRequestForTrack(t, m.cfg)
	m.qm.Enqueue(queue.Queued{Title: t.Name, RequestJSON: backend.DownloadRequestJSON(req)})
	m.home.input.Blur()
	m.home.inputFocused = false
	m.switchTab(3)
	return func() tea.Msg { return enqueuedMsg{1} }
}

func resolveCmd(query string) tea.Cmd {
	return func() tea.Msg {
		query = strings.TrimSpace(query)
		if query == "" {
			return resultsMsg{kind: "search"}
		}
		if strings.Contains(query, "://") || strings.HasPrefix(query, "www.") {
			r, err := backend.ResolveURL(query)
			if err != nil {
				return resultsMsg{err: err, kind: "URL"}
			}
			return resultsMsg{tracks: r.Tracks, kind: "URL"}
		}
		tracks, err := backend.Search(query, 20)
		return resultsMsg{tracks: tracks, err: err, kind: "search"}
	}
}

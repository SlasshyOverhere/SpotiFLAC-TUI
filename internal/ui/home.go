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
	if m.home.typing {
		b.WriteString(dimStyle.Render("URL or search query: "))
		b.WriteString(selStyle.Render(m.home.input + "▏"))
	} else {
		b.WriteString(dimStyle.Render("type 'e' to enter a URL or query · " +
			fmt.Sprintf("%d results", len(m.home.results))))
	}
	b.WriteString("\n\n")
	for i, t := range m.home.results {
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

func (m *Model) homeKey(msg tea.KeyMsg) tea.Cmd {
	if m.home.typing {
		if msg.Type == tea.KeyEsc {
			m.home.typing = false
			return nil
		}
		if msg.Type == tea.KeyEnter {
			query := m.home.input
			m.home.typing = false
			m.home.input = ""
			return resolveCmd(query)
		}
		if typeInto(&m.home.input, msg) {
			return nil
		}
		return nil
	}
	switch msg.Type {
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
			m.home.typing = true
		case "d":
			if len(m.home.results) > 0 {
				t := m.home.results[m.home.cursor]
				req := backend.DownloadRequestForTrack(t, m.cfg)
				m.qm.Enqueue(queue.Queued{Title: t.Name, RequestJSON: backend.DownloadRequestJSON(req)})
				return func() tea.Msg { return enqueuedMsg{1} }
			}
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
			return resultsMsg{tracks: r.Tracks, kind: "URL: " + r.Name}
		}
		tracks, err := backend.Search(query, 20)
		return resultsMsg{tracks: tracks, err: err, kind: "search"}
	}
}

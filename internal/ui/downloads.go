package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
)

func (m *Model) renderDownloads() string {
	items := m.qm.Items()
	if len(items) == 0 {
		return dimStyle.Render("queue is empty — download something from Home")
	}
	var b strings.Builder
	for i, it := range items {
		line := fmt.Sprintf("%s", it.Title)
		if i == m.dl.cursor {
			b.WriteString(selStyle.Render("▸ "))
		} else {
			b.WriteString("  ")
		}
		bar := progressBar(it.Progress.Progress)
		switch it.Status {
		case queue.StatusDone:
			b.WriteString(line + "  " + okStyle.Render("✓ done"))
			if it.Result != nil && it.Result.FilePath != "" {
				b.WriteString(dimStyle.Render("  " + it.Result.FilePath))
			}
		case queue.StatusError:
			b.WriteString(line + "  " + errStyle.Render("✗ "+it.Error))
		case queue.StatusCancelled:
			b.WriteString(line + "  " + dimStyle.Render("cancelled"))
		case queue.StatusRunning:
			stage := it.Progress.Stage
			if it.Progress.SpeedMBps > 0 {
				stage += fmt.Sprintf(" %.1f MB/s", it.Progress.SpeedMBps)
			}
			b.WriteString(line + "  " + bar + " " + fmt.Sprintf("%3.0f%%", it.Progress.Progress) + dimStyle.Render("  "+stage))
		default:
			b.WriteString(line + "  " + dimStyle.Render("queued"))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func progressBar(pct float64) string {
	const width = 20
	filled := int(pct / 100 * width)
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("·", width-filled) + "]"
}

func (m *Model) dlKey(msg tea.KeyMsg) tea.Cmd {
	items := m.qm.Items()
	if len(items) == 0 {
		return nil
	}
	switch msg.Type {
	case tea.KeyUp:
		if m.dl.cursor > 0 {
			m.dl.cursor--
		}
	case tea.KeyDown:
		if m.dl.cursor < len(items)-1 {
			m.dl.cursor++
		}
	case tea.KeyRunes:
		switch msg.String() {
		case "c":
			it := items[m.dl.cursor]
			if it.Status == queue.StatusQueued || it.Status == queue.StatusRunning {
				m.qm.Cancel(it.ID)
				m.msg = "cancelling " + it.Title
			}
		case "k":
			if m.dl.cursor > 0 {
				m.dl.cursor--
			}
		case "j":
			if m.dl.cursor < len(items)-1 {
				m.dl.cursor++
			}
		}
	}
	return nil
}

// ---- auth ----

func (m *Model) authCmd() tea.Cmd {
	return func() tea.Msg {
		reqs, err := backend.PendingAuthRequests()
		if err != nil || len(reqs) == 0 {
			if err != nil {
				m.msg = err.Error()
			} else {
				m.msg = "no pending auth requests"
			}
			return nil
		}
		m.auth = &authState{req: reqs[0]}
		return nil
	}
}

func (m *Model) renderAuth() string {
	var b strings.Builder
	b.WriteString(selStyle.Render("Verification required for " + m.auth.req.ExtensionID + "\n"))
	b.WriteString(dimStyle.Render("Open this URL in your browser and sign in:\n"))
	b.WriteString(m.auth.req.AuthURL + "\n\n")
	b.WriteString(dimStyle.Render("Paste the code/redirect URL: "))
	b.WriteString(m.auth.code + "▏")
	return b.String()
}

func (m *Model) handleAuthKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.auth = nil
	case tea.KeyEnter:
		backend.SetAuthCode(m.auth.req.ExtensionID, m.auth.code)
		m.auth = nil
		m.msg = "auth code submitted"
		return m, loadExts()
	default:
		typeInto(&m.auth.code, msg)
	}
	return m, nil
}

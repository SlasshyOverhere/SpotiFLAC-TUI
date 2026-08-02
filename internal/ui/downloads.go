package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/queue"
)

func (m *Model) renderDownloads() string {
	items := m.qm.Items()
	m.listY = 0
	if len(items) == 0 {
		return dimStyle.Render("queue is empty — download something from Home")
	}
	visible := clamp(m.h-10, 5, 30)
	start, end := m.listWindow(len(items), m.dl.cursor, visible)
	m.listWinStart = start
	var b strings.Builder
	for i := start; i < end; i++ {
		it := items[i]
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

func (m *Model) dlAction(msg tea.KeyMsg) tea.Cmd {
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

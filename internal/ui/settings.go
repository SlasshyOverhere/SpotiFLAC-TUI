package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type savedMsg struct{}

var qualities = []string{"lossless", "high", "normal"}

func (m *Model) renderSettings() string {
	var b strings.Builder
	if m.settings.typing {
		b.WriteString(dimStyle.Render("Output directory: "))
		b.WriteString(selStyle.Render(m.settings.input + "▏"))
	} else {
		b.WriteString(dimStyle.Render("output: " + m.cfg.OutputDir))
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("quality: "))
	b.WriteString(selStyle.Render(m.cfg.Quality))
	b.WriteString(dimStyle.Render("  (" + strings.Join(qualities, " / ") + " — toggle with q)\n"))
	b.WriteString(fmt.Sprintf("metadata: %v\n", m.cfg.EmbedMetadata))
	b.WriteString(fmt.Sprintf("lyrics:   %v\n", m.cfg.EmbedLyrics))
	b.WriteString(fmt.Sprintf("update check: %v\n", m.cfg.UpdateCheck))
	return b.String()
}

func (m *Model) settingsKey(msg tea.KeyMsg) tea.Cmd {
	if m.settings.typing {
		if msg.Type == tea.KeyEsc {
			m.settings.input = m.cfg.OutputDir
			m.settings.typing = false
			return nil
		}
		if msg.Type == tea.KeyEnter {
			m.cfg.OutputDir = m.settings.input
			m.settings.typing = false
			return m.saveCmd()
		}
		if typeInto(&m.settings.input, msg) {
			return nil
		}
		return nil
	}
	switch msg.Type {
	case tea.KeyRunes:
		switch msg.String() {
		case "e":
			m.settings.typing = true
		case "q":
			idx := indexOf(qualities, m.cfg.Quality)
			m.cfg.Quality = qualities[(idx+1)%len(qualities)]
		case "m":
			m.cfg.EmbedMetadata = !m.cfg.EmbedMetadata
		case "l":
			m.cfg.EmbedLyrics = !m.cfg.EmbedLyrics
		case "u":
			m.cfg.UpdateCheck = !m.cfg.UpdateCheck
		case "s":
			return m.saveCmd()
		}
	}
	return nil
}

func (m *Model) saveCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.cfg.Save(); err != nil {
			m.msg = err.Error()
			return nil
		}
		return savedMsg{}
	}
}

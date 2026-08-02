package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/queue"
)

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestTuiShell(t *testing.T) {
	var m tea.Model = New(app.Default(), queue.New())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if !strings.Contains(m.View(), "SpotiFLAC TUI") {
		t.Fatalf("header missing:\n%s", m.View())
	}

	for _, tab := range []string{"2", "3", "4", "5", "1"} {
		m, _ = m.Update(key(tab))
		if v := m.View(); strings.TrimSpace(v) == "" {
			t.Fatalf("tab %s renders empty", tab)
		}
	}
}

func TestTuiTyping(t *testing.T) {
	var m tea.Model = New(app.Default(), queue.New())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	m, _ = m.Update(key("e")) // enter typing on Home
	for _, r := range "spotify.com/track/xyz" {
		m, _ = m.Update(key(string(r)))
	}
	mm, ok := m.(*Model)
	if !ok {
		t.Fatalf("model type assertion failed")
	}
	if mm.home.input != "spotify.com/track/xyz" {
		t.Fatalf("typing broken: %q", mm.home.input)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if mm.home.typing {
		t.Fatalf("esc should stop typing")
	}
}

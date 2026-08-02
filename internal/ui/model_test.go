package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"spotiflac-cli/internal/app"
	"spotiflac-cli/internal/backend"
	"spotiflac-cli/internal/queue"
)

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestTuiShell(t *testing.T) {
	var m tea.Model = New(app.Default(), queue.New())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()
	if !strings.Contains(view, "SPOTIFLAC") {
		t.Fatalf("header missing:\n%s", m.View())
	}
	if !strings.Contains(view, "DOWNLOAD STUDIO") || !strings.Contains(view, "Home") {
		t.Fatalf("dashboard navigation missing:\n%s", view)
	}
	mm := m.(*Model)
	if mm.boxLeft <= 0 || mm.boxLeft >= 120 {
		t.Fatalf("sidebar geometry missing, got left=%d", mm.boxLeft)
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

	mm := m.(*Model)
	m, _ = m.Update(key("e")) // focus home input
	if !mm.home.inputFocused {
		t.Fatalf("e should focus the input")
	}
	for _, r := range "spotify.com/track/xyz" {
		m, _ = m.Update(key(string(r)))
	}
	if mm.home.input.Value() != "spotify.com/track/xyz" {
		t.Fatalf("typing broken: %q", mm.home.input.Value())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if mm.home.inputFocused {
		t.Fatalf("esc should blur the input")
	}
}

func TestTuiMouseWheel(t *testing.T) {
	var m tea.Model = New(app.Default(), queue.New())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	mm := m.(*Model)
	// fill some fake results
	for i := 0; i < 10; i++ {
		mm.home.results = append(mm.home.results, backend.Track{ID: "x", Name: "T"})
	}
	// wheel down
	mm.home.cursor = 0
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if mm.home.cursor != 1 {
		t.Fatalf("wheel down should move cursor, got %d", mm.home.cursor)
	}
}

func TestTuiFitsNarrowTerminal(t *testing.T) {
	var m tea.Model = New(app.Default(), queue.New())
	m, _ = m.Update(tea.WindowSizeMsg{Width: 52, Height: 14})
	mm := m.(*Model)
	mm.home.results = []backend.Track{{
		Name:      "A deliberately long title that must never escape the frame",
		Artists:   "An artist with a deliberately long name",
		AlbumName: "An album with enough metadata to expose wrapping bugs",
	}}
	view := m.View()
	for _, line := range strings.Split(view, "\n") {
		if width := lipgloss.Width(line); width > 52 {
			t.Fatalf("line exceeds terminal width: %d\n%s", width, line)
		}
	}
}

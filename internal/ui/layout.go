package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"spotiflac-cli/internal/queue"
)

var (
	dashboardBg    = lipgloss.Color("235")
	sidebarBg      = lipgloss.Color("233")
	panelBg        = lipgloss.Color("236")
	brandColor     = lipgloss.Color("81")
	activeColor    = lipgloss.Color("62")
	secondaryColor = lipgloss.Color("212")

	dashboardStyle = lipgloss.NewStyle().Background(dashboardBg)
	sidebarStyle   = lipgloss.NewStyle().Background(sidebarBg)
	panelStyle     = lipgloss.NewStyle().Background(panelBg)
	brandText      = lipgloss.NewStyle().Bold(true).Foreground(brandColor)
	mutedText      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	activeNav      = lipgloss.NewStyle().Background(activeColor).Foreground(lipgloss.Color("255")).Bold(true)
	keyText        = lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
)

func dashboardSidebarWidth(width int) int {
	if width < 48 {
		return max(width/3, 1)
	}
	return max(width/3, 24)
}

func (m *Model) renderDashboard() string {
	sidebarWidth := dashboardSidebarWidth(m.w)
	mainWidth := max(m.w-sidebarWidth, 1)
	m.boxLeft = sidebarWidth
	m.boxTop = 3

	content := m.content()
	panelHeight := max(m.h-4, 3)
	panel := renderPanel(tabNames[m.tab], content, mainWidth, panelHeight)

	main := make([]string, 0, m.h)
	main = append(main, dashboardLine(mainWidth, "  "+brandText.Render("SPOTIFLAC")+mutedText.Render("  /  "+tabNames[m.tab])))
	main = append(main, dashboardLine(mainWidth, "  "+mutedText.Render("lossless music workspace")+strings.Repeat(" ", max(mainWidth-31, 0))))
	main = append(main, dashboardLine(mainWidth, mutedText.Render(strings.Repeat("─", mainWidth))))
	main = append(main, panel...)
	main = append(main, dashboardLine(mainWidth, "  "+m.statusLine(mainWidth)))
	main = fitLines(main, m.h, mainWidth, dashboardStyle)

	sidebar := strings.Join(renderSidebar(sidebarWidth, m.h, m.tab, m.qm), "\n")
	mainView := strings.Join(main, "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainView)
}

func (m *Model) statusLine(width int) string {
	hint := m.renderFooter()
	if m.msg != "" {
		hint = keyText.Render("STATUS") + " " + m.msg + "   " + mutedText.Render("· "+m.hint())
	}
	return fitLine(hint, max(width-4, 1))
}

func (m *Model) hint() string {
	hints := map[int]string{
		0: "e search   j/k move   d download   a all",
		1: "r add repo   i install   R refresh",
		2: "x toggle   j/k reorder   m save   v verify",
		3: "j/k move   c cancel",
		4: "e edit   q quality   m metadata   l lyrics   s save",
	}
	return hints[m.tab]
}

func renderPanel(title, content string, width, height int) []string {
	inner := max(width-4, 1)
	rows := []string{panelLine(width, "╭─ "+brandText.Render(strings.ToUpper(title))+" "+strings.Repeat("─", max(inner-lipgloss.Width(title)-4, 0))+"╮")}
	rows = append(rows, fitBlock(content, inner, max(height-2, 1), panelStyle)...)
	rows = append(rows, panelLine(width, "╰"+strings.Repeat("─", max(width-2, 0))+"╯"))
	return fitLines(rows, height, width, panelStyle)
}

func renderSidebar(width, height, active int, qm *queue.Manager) []string {
	lines := make([]string, 0, height)
	lines = append(lines, sidebarLine(width, ""))
	lines = append(lines, sidebarLine(width, "  "+brandText.Render("S")))
	lines = append(lines, sidebarLine(width, "  "+brandText.Render("SpotiFLAC")))
	lines = append(lines, sidebarLine(width, "  "+mutedText.Render("DOWNLOAD STUDIO")))
	lines = append(lines, sidebarLine(width, ""))
	for i, name := range tabNames {
		label := fmt.Sprintf("  %d  %s", i+1, name)
		if i == active {
			lines = append(lines, activeNav.Render(fitLine(label, width)))
		} else {
			lines = append(lines, sidebarLine(width, mutedText.Render(label)))
		}
	}
	lines = append(lines, sidebarLine(width, ""))
	items := qm.Items()
	running, queued := 0, 0
	for _, item := range items {
		switch item.Status {
		case queue.StatusRunning:
			running++
		case queue.StatusQueued:
			queued++
		}
	}
	lines = append(lines, sidebarLine(width, "  "+keyText.Render("QUEUE")))
	lines = append(lines, sidebarLine(width, "  "+mutedText.Render(fmt.Sprintf("%d total   %d active", len(items), running))))
	if queued > 0 {
		lines = append(lines, sidebarLine(width, "  "+mutedText.Render(fmt.Sprintf("%d waiting", queued))))
	}

	for len(lines) < height-3 {
		lines = append(lines, sidebarLine(width, ""))
	}
	lines = append(lines, sidebarLine(width, "  "+keyText.Render("KEYS")))
	lines = append(lines, sidebarLine(width, "  "+mutedText.Render("tab / 1-5 navigate")))
	lines = append(lines, sidebarLine(width, "  "+mutedText.Render("q quit   ? help")))
	return fitLines(lines, height, width, sidebarStyle)
}

func panelLine(width int, value string) string {
	return panelStyle.Render(fitLine(value, width))
}

func sidebarLine(width int, value string) string {
	return sidebarStyle.Render(fitLine(value, width))
}

func dashboardLine(width int, value string) string {
	return dashboardStyle.Render(fitLine(value, width))
}

func fitBlock(content string, width, height int, style lipgloss.Style) []string {
	lines := strings.Split(strings.TrimSuffix(content, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	rows := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		rows = append(rows, style.Render(fitLine(line, width)))
	}
	return rows
}

func fitLines(lines []string, height, width int, style lipgloss.Style) []string {
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, style.Render(strings.Repeat(" ", width)))
	}
	for i, line := range lines {
		lines[i] = style.Render(fitLine(line, width))
	}
	return lines
}

func fitLine(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = ansi.Truncate(value, width, "…")
	return value + strings.Repeat(" ", max(width-lipgloss.Width(value), 0))
}

package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"spotiflac-cli/internal/backend"
)

type authPendingMsg struct {
	req backend.AuthRequest
	err error
}

func (m *Model) authCmd() tea.Cmd {
	return func() tea.Msg {
		reqs, err := backend.PendingAuthRequests()
		if err != nil {
			return authPendingMsg{err: err}
		}
		if len(reqs) == 0 {
			return authPendingMsg{}
		}
		return authPendingMsg{req: reqs[0]}
	}
}

func (m *Model) renderAuth() string {
	return selStyle.Render("Verification required for "+m.auth.req.ExtensionID+"\n") +
		dimStyle.Render("Open this URL in your browser and sign in:\n") +
		m.auth.req.AuthURL + "\n\n" +
		selStyle.Render("▸ ")+m.auth.input.View()
}

func (m *Model) updateAuth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.auth = nil
	case tea.KeyEnter:
		backend.SetAuthCode(m.auth.req.ExtensionID, m.auth.input.Value())
		m.auth = nil
		m.msg = "auth code submitted"
		return m, loadExts()
	default:
		var cmd tea.Cmd
		m.auth.input, cmd = m.auth.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"quasi.db_analysis_agent/internal/domain"
)

// switchSessionMsg se dispara al seleccionar una sesión de la lista.
type switchSessionMsg struct {
	sessionID string
}

// SessionListModel maneja la lista lateral de sesiones.
type SessionListModel struct {
	cursor int
}

func NewSessionListModel() SessionListModel {
	return SessionListModel{cursor: 0}
}

func (m SessionListModel) Init() tea.Cmd { return nil }

func (m SessionListModel) Update(msg tea.Msg, sm *domain.SessionManager) (SessionListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(sm.Sessions)-1 {
				m.cursor++
			}
		case "enter":
			ids := sessionIDs(sm)
			if m.cursor >= 0 && m.cursor < len(ids) {
				return m, func() tea.Msg {
					return switchSessionMsg{sessionID: ids[m.cursor]}
				}
			}
		}
	}
	return m, nil
}

func (m SessionListModel) View(sm *domain.SessionManager) string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render(" Sessions "))
	b.WriteString("\n\n")

	if len(sm.Sessions) == 0 {
		b.WriteString(SubtitleStyle.Render("No sessions. Press 'n' to create one."))
		return b.String()
	}

	ids := sessionIDs(sm)
	for i, id := range ids {
		s := sm.Sessions[id]
		line := fmt.Sprintf("[%s] %s", id, s.Name)
		if sm.ActiveID == id {
			line = SelectedSessionStyle.Render("▸ " + line)
		} else if i == m.cursor {
			line = SessionItemStyle.Render("> " + line)
		} else {
			line = SessionItemStyle.Render("  " + line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render("↑/↓ navigate • enter switch • n new"))
	return b.String()
}

func sessionIDs(sm *domain.SessionManager) []string {
	ids := make([]string, 0, len(sm.Sessions))
	for k := range sm.Sessions {
		ids = append(ids, k)
	}
	// ordenar alfabéticamente para estabilidad
	for i := 0; i < len(ids)-1; i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] > ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}

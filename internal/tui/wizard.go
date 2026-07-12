package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"quasi.db_analysis_agent/internal/domain"
)

// WizardModel es el submodelo para crear una nueva sesión de DB.
type WizardModel struct {
	inputs   []textinput.Model
	focusIdx int
}

func NewWizardModel() WizardModel {
	m := WizardModel{
		inputs: make([]textinput.Model, 5),
	}

	labels := []string{"Host", "Port", "User", "Password", "Database"}
	defaults := []string{"localhost", "3306", "", "", ""}

	for i := range m.inputs {
		t := textinput.New()
		t.Placeholder = labels[i]
		t.CharLimit = 64
		t.Width = 40
		if defaults[i] != "" {
			t.SetValue(defaults[i])
		}
		if i == 3 {
			t.EchoMode = textinput.EchoPassword
		}
		m.inputs[i] = t
	}
	m.inputs[0].Focus()
	return m
}

type createSessionMsg struct {
	config domain.ConnectionConfig
}

func (m WizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// updateInputs pasa el mensaje a TODOS los inputs, permitiendo que cada uno
// maneje su propio estado interno (buffer, cursor, etc.).
func (m *WizardModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

// setFocus aplica el foco visual al input activo y quita el foco del resto.
func (m *WizardModel) setFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focusIdx {
			cmds[i] = m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
			m.inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle()
			m.inputs[i].TextStyle = lipgloss.NewStyle()
		}
	}
	return tea.Batch(cmds...)
}

func (m WizardModel) Update(msg tea.Msg) (WizardModel, tea.Cmd) {
	// SIEMPRE pasar el mensaje a todos los inputs primero.
	// Esto permite que el input activo capture el texto escrito.
	cmd := m.updateInputs(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focusIdx < len(m.inputs)-1 {
				m.focusIdx++
				return m, tea.Batch(cmd, m.setFocus())
			}
			// Submit desde el último campo
			return m, tea.Batch(cmd, func() tea.Msg {
				return createSessionMsg{
					config: domain.ConnectionConfig{
						Type:     domain.MySQL,
						Host:     m.inputs[0].Value(),
						Port:     m.inputs[1].Value(),
						User:     m.inputs[2].Value(),
						Password: m.inputs[3].Value(),
						Database: m.inputs[4].Value(),
					},
				}
			})

		case tea.KeyShiftTab, tea.KeyUp:
			m.focusIdx--
			if m.focusIdx < 0 {
				m.focusIdx = len(m.inputs) - 1
			}
			return m, tea.Batch(cmd, m.setFocus())

		case tea.KeyTab, tea.KeyDown:
			m.focusIdx++
			if m.focusIdx >= len(m.inputs) {
				m.focusIdx = 0
			}
			return m, tea.Batch(cmd, m.setFocus())
		}
	}

	return m, cmd
}

func (m WizardModel) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle.Render(" New Session "))
	b.WriteString("\n\n")
	b.WriteString("Fill the connection details and press Enter to connect.\n\n")

	labels := []string{"Host    ", "Port    ", "User    ", "Password", "Database"}
	for i := range m.inputs {
		b.WriteString(fmt.Sprintf("%s  %s\n\n", labels[i], m.inputs[i].View()))
	}
	b.WriteString(SubtitleStyle.Render("tab/shift+tab to navigate • enter to submit"))
	return b.String()
}

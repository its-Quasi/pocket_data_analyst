package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"quasi.db_analysis_agent/internal/database"
	"quasi.db_analysis_agent/internal/domain"
	"quasi.db_analysis_agent/internal/llm"
	"quasi.db_analysis_agent/internal/runner"
)

type viewState int

const (
	viewSessionList viewState = iota
	viewWizard
	viewChat
)

// AppModel es el modelo raíz de la aplicación TUI.
type AppModel struct {
	width  int
	height int

	state      viewState
	sm         *domain.SessionManager
	llmClient  *llm.Client
	llmBaseURL string

	wizard      WizardModel
	sessionList SessionListModel
	input       textarea.Model
	spinner     spinner.Model
	vp          viewport.Model
	loading     bool
	agentStatus string
}

func NewAppModel(baseURL string) AppModel {
	ta := textarea.New()
	ta.Placeholder = "Ask something..."
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle()
	ta.BlurredStyle = ta.FocusedStyle
	ta.SetWidth(40)
	ta.SetHeight(1)
	ta.CharLimit = 0
	ta.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))

	return AppModel{
		state:       viewSessionList,
		sm:          domain.NewSessionManager(),
		llmClient:   llm.NewClient(baseURL),
		llmBaseURL:  baseURL,
		wizard:      NewWizardModel(),
		sessionList: NewSessionListModel(),
		input:       ta,
		spinner:     sp,
		vp:          viewport.New(80, 20),
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textarea.Blink)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizePanes()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.state {
		case viewSessionList:
			if msg.String() == "n" {
				m.state = viewWizard
				m.wizard = NewWizardModel()
				return m, m.wizard.Init()
			}
			var cmd tea.Cmd
			m.sessionList, cmd = m.sessionList.Update(msg, m.sm)
			return m, cmd

		case viewWizard:
			if msg.String() == "esc" {
				m.state = viewSessionList
				return m, nil
			}
			var cmd tea.Cmd
			m.wizard, cmd = m.wizard.Update(msg)
			return m, cmd

		case viewChat:
			// Scroll del viewport
			switch msg.Type {
			case tea.KeyUp:
				m.vp.LineUp(1)
				return m, nil
			case tea.KeyDown:
				m.vp.LineDown(1)
				return m, nil
			case tea.KeyPgUp:
				m.vp.HalfViewUp()
				return m, nil
			case tea.KeyPgDown:
				m.vp.HalfViewDown()
				return m, nil
			}

			if msg.String() == "esc" {
				m.state = viewSessionList
				m.input.SetValue("")
				m.input.SetHeight(1)
				return m, nil
			}

			// Enter envía, Shift+Enter inserta newline
			switch msg.String() {
			case "enter":
				if !m.loading {
					q := strings.TrimSpace(m.input.Value())
					if q != "" {
						m.loading = true
						m.agentStatus = "Generating..."
						session := m.sm.GetActive()
						session.Messages = append(session.Messages, domain.Message{
							Role:    domain.RoleUser,
							Content: q,
						})
						m.input.SetValue("")
						m.input.SetHeight(1)
						m.refreshViewport()
						m.vp.GotoBottom()
						return m, tea.Batch(
							m.spinner.Tick,
							StartAgent(m.llmClient, session),
						)
					}
				}
				return m, nil
			case "shift+enter":
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(tea.KeyMsg{Type: tea.KeyEnter})
				m.adjustInputHeight()
				return m, cmd
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				m.adjustInputHeight()
				return m, cmd
			}
		}

	case createSessionMsg:
		reader := database.MySQLReader{Config: msg.config}
		ddlInfo, err := reader.ReadDDL()
		if err != nil {
			m.state = viewSessionList
			return m, tea.Printf("Error connecting: %v", err)
		}

		session := &domain.Session{
			ID:      m.sm.NextID(),
			Name:    msg.config.Database,
			Config:  msg.config,
			DDLInfo: *ddlInfo,
		}
		m.sm.AddSession(session)
		m.state = viewSessionList
		return m, nil

	case switchSessionMsg:
		m.sm.SwitchSession(msg.sessionID)
		m.state = viewChat
		m.input.SetValue("")
		m.input.SetHeight(1)
		m.input.Focus()
		m.refreshViewport()
		return m, nil

	case agentCodeMsg:
		// Código generado o reparado — mostrarlo y ejecutarlo
		m.agentStatus = "Executing..."
		m.refreshViewport()
		m.vp.GotoBottom()

		return m, executeCode(msg.code, msg.explanation, msg.attempt)

	case agentExecMsg:
		session := m.sm.GetActive()
		if session == nil {
			m.loading = false
			m.agentStatus = ""
			return m, nil
		}

		if msg.execErr == nil {
			// ÉXITO — actualizar el último mensaje assistant con el output
			lastIdx := len(session.Messages) - 1
			if lastIdx >= 0 && session.Messages[lastIdx].Role == domain.RoleAssistant {
				session.Messages[lastIdx].Content = msg.output

				chartPath := extractChartPath(msg.output)
				if chartPath != "" {
					_ = runner.OpenBrowser(chartPath)
					if session.Messages[lastIdx].Explanation != "" {
						session.Messages[lastIdx].Explanation += fmt.Sprintf("\n\nChart opened: %s", chartPath)
					} else {
						session.Messages[lastIdx].Explanation = fmt.Sprintf("Chart opened: %s", chartPath)
					}
				}
			}
			m.loading = false
			m.agentStatus = ""
			m.refreshViewport()
			m.vp.GotoBottom()
			return m, nil
		}

		// FALLO — marcar el último mensaje assistant como fallido
		lastIdx := len(session.Messages) - 1
		if lastIdx >= 0 && session.Messages[lastIdx].Role == domain.RoleAssistant {
			session.Messages[lastIdx].Failed = true
			session.Messages[lastIdx].Content = msg.output
		}
		m.refreshViewport()
		m.vp.GotoBottom()

		// ¿Quedan reintentos?
		if msg.attempt+1 >= maxRetries {
			m.loading = false
			m.agentStatus = ""
			session.Messages = append(session.Messages, domain.Message{
				Role:    domain.RoleAssistant,
				Content: fmt.Sprintf("Failed after %d attempts. Could not generate working code.", maxRetries),
			})
			m.refreshViewport()
			m.vp.GotoBottom()
			return m, nil
		}

		// Iniciar reparación
		m.agentStatus = fmt.Sprintf("Repairing (attempt %d/%d)...", msg.attempt+2, maxRetries)
		return m, repairCode(m.llmClient, session, msg.code, msg.output, msg.attempt+1)

	case agentDoneMsg:
		m.loading = false
		m.agentStatus = ""
		if msg.err != nil {
			session := m.sm.GetActive()
			if session != nil {
				session.Messages = append(session.Messages, domain.Message{
					Role:    domain.RoleAssistant,
					Content: fmt.Sprintf("Error: %v", msg.err),
				})
			}
		}
		m.refreshViewport()
		m.vp.GotoBottom()
		return m, nil
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	switch m.state {
	case viewSessionList:
		return m.sessionList.View(m.sm)
	case viewWizard:
		return BoxStyle.Render(m.wizard.View())
	case viewChat:
		return m.chatLayoutView()
	}
	return ""
}

func (m *AppModel) adjustInputHeight() {
	lines := strings.Count(m.input.Value(), "\n") + 1
	if lines < 1 {
		lines = 1
	}
	if lines > 6 {
		lines = 6
	}
	m.input.SetHeight(lines)
	m.resizePanes()
}

func (m AppModel) innerRightWidth() int {
	leftW := m.width / 4
	if leftW < 20 {
		leftW = 20
	}
	if leftW > 35 {
		leftW = 35
	}
	inner := m.width - leftW - 6
	if inner < 10 {
		inner = 10
	}
	return inner
}

func (m *AppModel) resizePanes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	innerRightW := m.innerRightWidth()

	// Reservar espacio para la altura MÁXIMA del input (6 líneas) + hint
	// Así el viewport tiene altura fija y no se mueve cuando el input crece
	maxInputHeight := 6
	hintHeight := 1
	vpHeight := m.height - maxInputHeight - hintHeight - 4
	if vpHeight < 5 {
		vpHeight = 5
	}

	m.vp.Width = innerRightW
	m.vp.Height = vpHeight

	m.input.SetWidth(innerRightW)
}

func (m AppModel) chatLayoutView() string {
	leftW := m.width / 4
	if leftW < 20 {
		leftW = 20
	}
	if leftW > 35 {
		leftW = 35
	}

	leftPane := LeftPaneStyle.
		Width(leftW).
		Height(m.height - 2).
		Render(m.sessionList.View(m.sm))

	rightContent := m.rightPaneView()
	rightPane := RightPaneStyle.
		Width(m.width - leftW - 2).
		Height(m.height - 2).
		Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

func (m AppModel) rightPaneView() string {
	session := m.sm.GetActive()
	if session == nil {
		return "No active session"
	}

	innerRightW := m.innerRightWidth()

	var b strings.Builder
	b.WriteString(TitleStyle.Render(fmt.Sprintf(" %s ", session.Name)))
	b.WriteString("\n")
	b.WriteString(m.vp.View())
	b.WriteString("\n")

	// Input box
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(innerRightW).
		Render(m.input.View())
	b.WriteString(inputBox)
	b.WriteString("\n")

	if m.loading {
		status := m.agentStatus
		if status == "" {
			status = "thinking..."
		}
		b.WriteString(m.spinner.View() + " " + status)
	} else {
		model := m.llmClient.Model()
		hint := fmt.Sprintf("%s  •  ↑/↓ scroll • esc back • enter send", model)
		b.WriteString(SubtitleStyle.Render(hint))
	}
	return b.String()
}

func (m *AppModel) refreshViewport() {
	session := m.sm.GetActive()
	if session == nil {
		m.vp.SetContent("")
		return
	}

	var b strings.Builder
	for _, msg := range session.Messages {
		if msg.IsError {
			continue
		}

		switch msg.Role {
		case domain.RoleUser:
			b.WriteString(UserMsgStyle.Render("You: "))
			b.WriteString(msg.Content)
			b.WriteString("\n\n")
		case domain.RoleAssistant:
			if msg.Failed {
				b.WriteString(FailedLabelStyle.Render("AI: ✗ FAILED"))
				b.WriteString("\n")
				b.WriteString(FailedCodeStyle.Render(msg.RawCode))
				b.WriteString("\n")
				if msg.Content != "" {
					b.WriteString(ErrorStyle.Render(msg.Content))
				}
				b.WriteString("\n\n")
			} else if msg.RawCode != "" && msg.Explanation != "" {
				b.WriteString(AssistantMsgStyle.Render("AI: "))
				b.WriteString("\n")
				b.WriteString(CodeBlockStyle.Render(msg.RawCode))
				b.WriteString("\n\n")
				b.WriteString(ExplanationStyle.Render(msg.Explanation))
				b.WriteString("\n\n")
			} else if msg.RawCode != "" {
				b.WriteString(AssistantMsgStyle.Render("AI: "))
				b.WriteString("\n")
				b.WriteString(CodeBlockStyle.Render(msg.RawCode))
				b.WriteString("\n")
				if msg.Content != "" {
					b.WriteString(ExplanationStyle.Render(msg.Content))
				}
				b.WriteString("\n\n")
			} else {
				b.WriteString(AssistantMsgStyle.Render("AI: "))
				b.WriteString(msg.Content)
				b.WriteString("\n\n")
			}
		}
	}
	m.vp.SetContent(b.String())
}

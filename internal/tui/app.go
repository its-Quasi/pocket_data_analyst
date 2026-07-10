package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
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

	wizard             WizardModel
	sessionList        SessionListModel
	input              textinput.Model
	spinner            spinner.Model
	vp                 viewport.Model
	loading            bool
}

// llmResponseMsg se devuelve desde la goroutine del AgentLoop.
type llmResponseMsg struct {
	code   string
	output string
	err    error
}

func NewAppModel(baseURL string) AppModel {
	in := textinput.New()
	in.Placeholder = "Ask something..."
	in.Width = 40
	in.Focus()

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
		input:       in,
		spinner:     sp,
		vp:          viewport.New(80, 20),
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick)
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
			return m, nil
		}

			if msg.Type == tea.KeyEnter && !m.loading {
				q := strings.TrimSpace(m.input.Value())
				if q != "" {
					m.loading = true
					session := m.sm.GetActive()
					session.Messages = append(session.Messages, domain.Message{
						Role:    domain.RoleUser,
						Content: q,
					})
					m.input.SetValue("")
					m.refreshViewport()
					m.vp.GotoBottom()
					return m, tea.Batch(
						m.spinner.Tick,
						AgentLoop(m.llmClient, session),
					)
				}
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
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
		m.input.Focus()
		m.refreshViewport()
		return m, nil

	case llmResponseMsg:
		m.loading = false
		if session := m.sm.GetActive(); session != nil {
			m.refreshViewport()
			m.vp.GotoBottom()
		}
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

func (m *AppModel) resizePanes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	leftW := m.width / 4
	if leftW < 20 {
		leftW = 20
	}
	if leftW > 35 {
		leftW = 35
	}

	rightW := m.width - leftW - 2
	if rightW < 10 {
		rightW = 10
	}

	inputHeight := 3
	vpHeight := m.height - inputHeight - 4
	if vpHeight < 5 {
		vpHeight = 5
	}

	m.vp.Width = rightW - 4
	m.vp.Height = vpHeight

	m.input.Width = rightW - 4
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

	var b strings.Builder
	b.WriteString(TitleStyle.Render(fmt.Sprintf(" %s ", session.Name)))
	b.WriteString("\n")
	b.WriteString(m.vp.View())
	b.WriteString("\n")
	b.WriteString(m.input.View())
	b.WriteString("\n")
	if m.loading {
		b.WriteString(m.spinner.View() + " thinking...")
	} else {
		b.WriteString(SubtitleStyle.Render("↑/↓ scroll • esc back • enter send"))
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
		switch msg.Role {
		case domain.RoleUser:
			b.WriteString(UserMsgStyle.Render("You: "))
			b.WriteString(msg.Content)
		case domain.RoleAssistant:
			b.WriteString(AssistantMsgStyle.Render("AI: "))
			b.WriteString(msg.Content)
		}
		b.WriteString("\n\n")
	}
	m.vp.SetContent(b.String())
}

// AgentLoop ejecuta el ciclo CodeAct: genera código, lo ejecuta,
// y si falla inyecta automáticamente el error como feedback al LLM
// para reintentar (máximo 3 veces). Muta directamente session.Messages.
func AgentLoop(client *llm.Client, session *domain.Session) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		maxRetries := 3

		for attempt := 0; attempt < maxRetries; attempt++ {
			systemPrompt := llm.BuildSystemPrompt(&session.DDLInfo, session.Config.DSN())
			msgs := llm.SessionMessagesToOpenAI(session.Messages, systemPrompt)

			generatedCode, err := client.Chat(ctx, msgs)
			if err != nil {
				session.Messages = append(session.Messages, domain.Message{
					Role:    domain.RoleAssistant,
					Content: fmt.Sprintf("LLM communication error: %v", err),
				})
				return llmResponseMsg{code: "", output: "", err: err}
			}

			output, execErr := runner.ExecuteTemporal(generatedCode)

			if execErr == nil {
				// ÉXITO TÉCNICO
				session.Messages = append(session.Messages, domain.Message{
					Role:    domain.RoleAssistant,
					Content: fmt.Sprintf("=== Generated Code ===\n%s\n\n=== Execution Result ===\n%s", generatedCode, output),
					RawCode: generatedCode,
				})
				return llmResponseMsg{code: generatedCode, output: output, err: nil}
			}

			// ERROR: guardar código que falló y luego el feedback automático
			session.Messages = append(session.Messages, domain.Message{
				Role:    domain.RoleAssistant,
				Content: fmt.Sprintf("=== Generated Code (attempt %d) ===\n%s", attempt+1, generatedCode),
				RawCode: generatedCode,
			})
			session.Messages = append(session.Messages, domain.Message{
				Role:    domain.RoleUser,
				Content: fmt.Sprintf("The previous code failed with error: %v\nOutput: %s\nPlease fix and regenerate.", execErr, output),
				IsError: true,
			})
		}

		// Agotó los 3 intentos sin éxito
		session.Messages = append(session.Messages, domain.Message{
			Role:    domain.RoleAssistant,
			Content: fmt.Sprintf("Failed after %d attempts. Could not generate working code.", maxRetries),
		})
		return llmResponseMsg{code: "", output: "", err: fmt.Errorf("failed after %d attempts", maxRetries)}
	}
}

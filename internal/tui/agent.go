package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"quasi.db_analysis_agent/internal/domain"
	"quasi.db_analysis_agent/internal/llm"
	"quasi.db_analysis_agent/internal/runner"
)

const maxRetries = 5

// --- Messages ---

// agentCodeMsg: el LLM generó código (inicial o reparado), listo para ejecutar.
type agentCodeMsg struct {
	code        string
	explanation string
	attempt     int
}

// agentExecMsg: la ejecución del código terminó (éxito o fracaso).
type agentExecMsg struct {
	code        string
	explanation string
	output      string
	execErr     error
	attempt     int
}

// agentDoneMsg: el ciclo completo terminó (éxito o error irrecuperable).
type agentDoneMsg struct {
	err error
}

// --- Commands ---

// StartAgent inicia el ciclo CodeAct: generar → ejecutar → reparar si falla.
func StartAgent(client *llm.Client, session *domain.Session) tea.Cmd {
	return generateCode(client, session, 0)
}

// generateCode llama al LLM con el historial de la sesión para generar código inicial.
func generateCode(client *llm.Client, session *domain.Session, attempt int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		systemPrompt := llm.BuildSystemPrompt(&session.DDLInfo, session.Config.DSN())
		msgs := llm.SessionMessagesToOpenAI(session.Messages, systemPrompt)

		resp, err := client.Chat(ctx, msgs)
		if err != nil {
			return agentDoneMsg{err: fmt.Errorf("LLM communication error: %w", err)}
		}

		code, explanation := parseLLMResponse(resp)

		session.Messages = append(session.Messages, domain.Message{
			Role:        domain.RoleAssistant,
			RawCode:     code,
			Explanation: explanation,
		})

		return agentCodeMsg{code: code, explanation: explanation, attempt: attempt}
	}
}

// executeCode ejecuta el código generado y devuelve el resultado.
func executeCode(code, explanation string, attempt int) tea.Cmd {
	return func() tea.Msg {
		output, execErr := runner.ExecuteTemporal(code)
		return agentExecMsg{
			code:        code,
			explanation: explanation,
			output:      output,
			execErr:     execErr,
			attempt:     attempt,
		}
	}
}

// repairCode construye un prompt de reparación autónomo y llama al LLM.
// No depende del historial: incluye el código fallido, el error y la
// petición original del usuario. Si el error es de go-echarts, inyecta
// código de referencia de la librería local.
func repairCode(client *llm.Client, session *domain.Session, failedCode, output string, attempt int) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		originalRequest := getOriginalRequest(session)

		// referenceCode := ""
		// if docssearch.IsChartRelated(output) {
		// 	searcher := docssearch.DefaultSearcher()
		// 	if searcher != nil {
		// 		referenceCode = searcher.SearchForError(output)
		// 	}
		// }

		repairPrompt := llm.BuildRepairPrompt(originalRequest, failedCode, output, "")
		msgs := llm.BuildRepairMessages(&session.DDLInfo, session.Config.DSN(), repairPrompt)

		resp, err := client.Chat(ctx, msgs)
		if err != nil {
			return agentDoneMsg{err: fmt.Errorf("LLM communication error: %w", err)}
		}

		code, explanation := parseLLMResponse(resp)

		session.Messages = append(session.Messages, domain.Message{
			Role:        domain.RoleAssistant,
			RawCode:     code,
			Explanation: explanation,
		})

		return agentCodeMsg{code: code, explanation: explanation, attempt: attempt}
	}
}

// --- Helpers ---

// getOriginalRequest busca la petición original del usuario más reciente
// (el último mensaje user que no sea feedback de error).
func getOriginalRequest(session *domain.Session) string {
	for i := len(session.Messages) - 1; i >= 0; i-- {
		if session.Messages[i].Role == domain.RoleUser && !session.Messages[i].IsError {
			return session.Messages[i].Content
		}
	}
	return ""
}

// extractChartPath busca la última línea no vacía del output que termine en .html.
func extractChartPath(output string) string {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && strings.HasSuffix(line, ".html") {
			return line
		}
	}
	return ""
}

// parseLLMResponse separa el código y la explicación de la respuesta del LLM.
func parseLLMResponse(response string) (code string, explanation string) {
	codeMarker := "---CODE---"
	explanationMarker := "---EXPLANATION---"

	codeIdx := strings.Index(response, codeMarker)
	explanationIdx := strings.Index(response, explanationMarker)

	if codeIdx == -1 || explanationIdx == -1 {
		return strings.TrimSpace(response), ""
	}

	code = strings.TrimSpace(response[codeIdx+len(codeMarker) : explanationIdx])
	explanation = strings.TrimSpace(response[explanationIdx+len(explanationMarker):])

	return code, explanation
}

package llm

import (
	"github.com/openai/openai-go/v3"
	"quasi.db_analysis_agent/internal/domain"
)

// BuildSystemPrompt genera el system prompt que incluye DDL y DSN.
func BuildSystemPrompt(ddl *domain.DDLInfo, dsn string) string {
	return `You are a Go code generation engine specialized in database queries.
You have access to the following MySQL database schema:

` + ddl.ToContextString() + `

Use this exact DSN for all database connections:
"` + dsn + `"

Output requirements:
- Output ONLY raw Go source code.
- The first line MUST be: package main
- The code MUST contain a func main().
- Do not output markdown.
- Do not output explanations.
- Do not output notes.
- Do not output examples.
- Do not output any text before or after the Go source code.
- The generated code must compile successfully with Go.
- When generating code that queries the database, use the exact DSN provided above.
`
}

// SessionMessagesToOpenAI convierte los mensajes internos de una sesión al formato de la librería.
func SessionMessagesToOpenAI(msgs []domain.Message, systemPrompt string) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs)+1)
	out = append(out, openai.SystemMessage(systemPrompt))
	for _, m := range msgs {
		switch m.Role {
		case domain.RoleUser:
			out = append(out, openai.UserMessage(m.Content))
		case domain.RoleAssistant:
			out = append(out, openai.AssistantMessage(m.Content))
		}
	}
	return out
}

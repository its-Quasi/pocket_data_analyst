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

About code execution:
- For ALL database query errors, use log.Fatal(err) instead of log.Printf or graceful returns.
- The program MUST exit with a non-zero status if the query fails or produces unexpected errors.
- Do not wrap query errors in if err != nil { log.Printf(...); return }.

About charts and visualizations:
- Never invent imports.
- Only import packages that exist.
- For go-echarts, the allowed imports are packages inside of github.com/go-echarts/go-echarts/v2 like charts, opts, etc...
- Do not import the root module.
- If the user asks for a chart, graph, visualization, plot, or any graphical representation.
- Available types: Bar (comparisons), Line (trends over time), Pie (proportions), Scatter (distributions). Choose the most appropriate based on the data.
- Create the directory charts/ inside already existing directory sandbox_area with os.MkdirAll if needed.
- Write the generated HTML to ./charts/ with a unique filename (e.g., chart_<unixtime>.html).
- Print the absolute file path of the HTML file as the LAST line of stdout using fmt.Println(path).
- Example:
  bar := charts.NewBar()
  bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Sales"}))
  bar.AddSeries("Revenue", values).SetXAxis(labels)
  os.MkdirAll("./sandbox_area/charts", 0755)
  f, _ := os.Create("./sandbox_area/charts/chart_123456.html")
  bar.Render(f)
  fmt.Println("./sandbox_area/charts/chart_123456.html")
  the location of the html file will be inside of directory sandbox_area/charts

Output format requirements:
You MUST output your response in the following EXACT format:

---CODE---
[Your Go source code here]
---EXPLANATION---
[A brief explanation in the user's language of what the code does and what results to expect]

Rules:
- The first section MUST be between ---CODE--- and ---EXPLANATION--- markers.
- The second section MUST be between ---EXPLANATION--- and the end of your response.
- The code section MUST contain ONLY raw Go source code, starting with "package main".
- The explanation section MUST be a natural language explanation of what the code does and what results the user should expect to see.
- Do NOT describe the code structure or implementation details in the explanation.
- DO describe what the query retrieves, what the output will show, and what insights can be drawn from the results.
- If a chart is generated, mention the chart path in the explanation.
- Do not output anything before ---CODE--- or after the explanation.
- The generated code must compile successfully with Go.
- When generating code that queries the database, use the exact DSN provided above.
`
}

// SessionMessagesToOpenAI convierte los mensajes internos de una sesión al formato de la librería.
// Cuando un mensaje del assistant contiene RawCode, se envía SOLO el código crudo
// para evitar que el LLM aprenda a generar markdown o texto decorativo.
func SessionMessagesToOpenAI(msgs []domain.Message, systemPrompt string) []openai.ChatCompletionMessageParamUnion {
	out := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs)+1)
	out = append(out, openai.SystemMessage(systemPrompt))
	for _, m := range msgs {
		switch m.Role {
		case domain.RoleUser:
			out = append(out, openai.UserMessage(m.Content))
		case domain.RoleAssistant:
			if m.RawCode != "" {
				out = append(out, openai.AssistantMessage(m.RawCode))
			} else {
				out = append(out, openai.AssistantMessage(m.Content))
			}
		}
	}
	return out
}

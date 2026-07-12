package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// Client es un wrapper sobre el cliente de OpenAI (Ollama compatible).
type Client struct {
	inner openai.Client
	model string
}

// NewClient crea un nuevo cliente apuntando a la URL del provider.
func NewClient(baseURL string) *Client {
	return &Client{
		inner: openai.NewClient(
			option.WithBaseURL(baseURL),
		),
		model: "gpt-oss:20b-cloud",
	}
}

// SetModel permite cambiar el modelo por defecto.
func (c *Client) SetModel(m string) {
	c.model = m
}

// Model devuelve el nombre del modelo activo.
func (c *Client) Model() string {
	return c.model
}

// Chat envía mensajes al LLM y devuelve la respuesta.
func (c *Client) Chat(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	res, err := c.inner.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf("llm chat error: %w", err)
	}
	return res.Choices[0].Message.Content, nil
}

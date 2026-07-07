package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	ctx := context.Background()
	godotenv.Load()

	client := openai.NewClient(
		option.WithBaseURL("http://localhost:11434/v1/"),
	)

	model := "gemma4:cloud"

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(
			`You are a Go code generation engine.
			Output requirements:
			- Output ONLY raw Go source code.
			- The first line MUST be: package main
			- The code MUST contain a func main().
			- Do not output markdown.
			- Do not output explanations.
			- Do not output notes.
			- Do not output examples.
			- Do not output any text before or after the Go source code.
			- The generated code must compile successfully with Go.`,
		),
	}

	input := catchUserInput()

	messages = append(messages, openai.UserMessage(input))
	res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Validando response de Ollama")
	response := res.Choices[0].Message.Content
	fmt.Println(response)
	ExecuteTemporal(response)
}

func ExecuteTemporal(gocode string) {
	root_path, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	sandbox_path := filepath.Join(root_path, "sandbox_area", "temporal.go")
	fmt.Println(sandbox_path)
	err = os.WriteFile(sandbox_path, []byte(gocode), 0644)

	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "run", sandbox_path)
	output, err := cmd.Output()

	if err != nil {
		panic(err)
	}

	fmt.Println(string(output))
	err = os.Remove(sandbox_path)

	if err != nil {
		panic(err)
	}
}

func catchUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	input := ""
	for {
		fmt.Print("Enter ur go ask> ")

		if !scanner.Scan() {
			break
		}

		input = strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input != "" {
			break
		}
	}
	return input
}

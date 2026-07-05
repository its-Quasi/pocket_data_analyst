package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {

	MOCK := `
	package main
	import (
		"fmt"
	)

	func main() {
		fmt.Println("Hello World!!!")
	}
	`

	root_path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sandbox_path := filepath.Join(root_path, "sandbox_area", "temporal.go")
	fmt.Println(sandbox_path)
	err = os.WriteFile(sandbox_path, []byte(MOCK), 0644)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "run", sandbox_path)
	// cmd.CombinedOutput()
	// err = cmd.Run()
	// if err != nil {
	// 	fmt.Println(err)
	// 	panic(err)
	// }
	output, err := cmd.Output()

	if err != nil {
		panic(err)
	}

	fmt.Println(string(output))
	// fmt.Println(cmd.Stdout)
	// ctx := context.Background()
	// godotenv.Load()

	// client := openai.NewClient(
	// 	option.WithBaseURL("http://localhost:11434/v1/"),
	// )

	// model := "gemma4:cloud"

	// messages := []openai.ChatCompletionMessageParamUnion{
	// 	openai.SystemMessage("You are a helpful assistant."),
	// }
	// scanner := bufio.NewScanner(os.Stdin)

	// for {
	// 	fmt.Print("\n> ")

	// 	if !scanner.Scan() {
	// 		break
	// 	}

	// 	input := strings.TrimSpace(scanner.Text())
	// 	if input == "" {
	// 		continue
	// 	}

	// 	messages = append(messages, openai.UserMessage(input))

	// 	res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
	// 		Model:    model,
	// 		Messages: messages,
	// 	})
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	fmt.Println(res.Choices[0].Message.Content)
	// 	messages = append(messages, res.Choices[0].Message.ToParam())
	// }

}

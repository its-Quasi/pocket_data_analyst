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

	"quasi.db_analysis_agent/internal/db"
)

func main() {
	ctx := context.Background()
	godotenv.Load()

	// 1. Seleccionar tipo de base de datos (por ahora solo MySQL)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("=== Database Connection ===")
	fmt.Println("Supported databases: mysql")
	fmt.Print("Enter database type (default: mysql): ")
	dbType, _ := reader.ReadString('\n')
	dbType = strings.TrimSpace(strings.ToLower(dbType))
	if dbType == "" {
		dbType = "mysql"
	}

	if dbType != "mysql" {
		fmt.Println("Only 'mysql' is supported at this time.")
		os.Exit(1)
	}

	// 2. Pedir credenciales
	fmt.Print("Host (default: localhost): ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}

	fmt.Print("Port (default: 3306): ")
	port, _ := reader.ReadString('\n')
	port = strings.TrimSpace(port)
	if port == "" {
		port = "3306"
	}

	fmt.Print("User: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	if user == "" {
		user = "root"
	}

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if password == "" {
		password = "root"
	}

	fmt.Print("Database name: ")
	database, _ := reader.ReadString('\n')
	database = strings.TrimSpace(database)
	if database == "" {
		database = "employees"
	}

	if user == "" || database == "" {
		fmt.Println("User and Database name are required.")
		os.Exit(1)
	}

	// 3. Conectar y leer DDL
	fmt.Println("\nConnecting to database and reading DDL...")
	config := db.ConnectionConfig{
		Type:     db.MySQL,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}

	mysqlReader := db.MySQLReader{Config: config}
	ddlInfo, err := mysqlReader.ReadDDL()
	if err != nil {
		fmt.Printf("Error reading DDL: %v\n", err)
		os.Exit(1)
	}

	ddlContext := ddlInfo.ToContextString()
	fmt.Printf("Loaded DDL for %d tables.\n\n", len(ddlInfo.Tables))

	// 4. Preparar el cliente LLM con el DDL y el DSN real en el system prompt
	actualDSN := config.DSN()

	client := openai.NewClient(
		option.WithBaseURL("http://localhost:11434/v1/"),
	)

	model := "gemma4:cloud"

	systemPrompt := `You are a Go code generation engine specialized in database queries.
You have access to the following MySQL database schema:

` + ddlContext + `

Use this exact DSN for all database connections:
"` + actualDSN + `"

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

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	// 5. Loop de interacción con el usuario
	for {
		input := catchUserInput()
		if input == "exit" || input == "quit" {
			fmt.Println("Bye!")
			break
		}

		messages = append(messages, openai.UserMessage(input))
		res, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    model,
			Messages: messages,
		})
		if err != nil {
			fmt.Printf("Error from LLM: %v\n", err)
			continue
		}

		response := res.Choices[0].Message.Content
		fmt.Println("=== Generated Code ===")
		fmt.Println(response)
		fmt.Println("=== Execution Result ===")
		ExecuteTemporal(response)
		fmt.Println()
	}
}

func ExecuteTemporal(gocode string) {
	root_path, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting wd: %v\n", err)
		return
	}

	sandbox_path := filepath.Join(root_path, "sandbox_area", "temporal.go")
	if err := os.WriteFile(sandbox_path, []byte(gocode), 0644); err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}

	cmd := exec.Command("go", "run", sandbox_path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Execution error: %v\nOutput: %s\n", err, string(output))
		return
	}

	fmt.Println(string(output))
	if err := os.Remove(sandbox_path); err != nil {
		fmt.Printf("Error removing file: %v\n", err)
	}
}

func catchUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter ur go ask> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			return input
		}
	}
	return ""
}

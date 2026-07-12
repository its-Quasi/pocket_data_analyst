package llm

import (
	"os"
	"path/filepath"
	"strings"
)

func CheckIntoDocs(files []string) string {
	var ref strings.Builder
	root := "internal/lib/go-echarts"
	for _, file := range files {
		path := filepath.Join(
			root,
			file,
		)

		// writeDebug("=== PATH ===\n" + path + "\n")
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// writeDebug("=== CONTENT ===\n" + string(content) + "\n")
		ref.WriteString("\n\n")
		ref.WriteString("SOURCE FILE: ")
		ref.WriteString(file)
		ref.WriteString("\n")
		ref.Write(content)
	}
	return ref.String()
}

// writeDebug escribe texto al archivo de debug sandbox_area/debug_docs.txt.
// Si el archivo no existe lo crea; si existe, agrega al final (append).
func writeDebug(text string) {
	debugPath := filepath.Join("sandbox_area", "debug_docs.txt")
	f, err := os.OpenFile(debugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(text)
}

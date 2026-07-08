package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ExecuteTemporal escribe el código Go en un archivo temporal, lo ejecuta y limpia.
func ExecuteTemporal(gocode string) (string, error) {
	rootPath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting wd: %w", err)
	}

	sandboxDir := filepath.Join(rootPath, "sandbox_area")
	if err := os.MkdirAll(sandboxDir, 0755); err != nil {
		return "", fmt.Errorf("creating sandbox dir: %w", err)
	}

	sandboxPath := filepath.Join(sandboxDir, "temporal.go")
	if err := os.WriteFile(sandboxPath, []byte(gocode), 0644); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	cmd := exec.Command("go", "run", sandboxPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("execution error: %w", err)
	}

	if rmErr := os.Remove(sandboxPath); rmErr != nil {
		return string(output), fmt.Errorf("cleanup error: %w", rmErr)
	}

	return string(output), nil
}

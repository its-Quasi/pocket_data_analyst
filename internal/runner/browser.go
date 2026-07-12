package runner

import (
	"os/exec"
	"runtime"
)

// OpenBrowser abre la ruta proporcionada en el navegador por defecto del sistema.
func OpenBrowser(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

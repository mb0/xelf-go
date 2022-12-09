package xps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GoTool(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	return cmd.Output()
}

func GoModPath(dir string) (string, error) {
	b, err := GoTool(dir, "list", "-m")
	if err != nil {
		return "", fmt.Errorf("go mod path for %s: %v", dir, err)
	}
	return strings.TrimSpace(string(b)), nil
}

func HasSource(m Manifest) bool {
	dir, _ := filepath.Split(m.Path)
	_, err := os.Stat(filepath.Join(dir, "plugin.go"))
	return err == nil
}
func Rebuild(m Manifest) error {
	dir, _ := filepath.Split(m.Path)
	res, err := GoTool(dir, "build", "-buildmode=plugin")
	if err != nil {
		return fmt.Errorf("%w %s", err, res)
	}
	return nil
}

package runner

import (
	"bytes"
	"os/exec"
	"strings"
)

func TerraformVersion(path string) (string, error) {
	cmd := exec.Command(path, "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.Trim(out.String(), "\n"), nil
}

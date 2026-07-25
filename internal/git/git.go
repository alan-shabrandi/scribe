package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute git diff command: %w", err)
	}

	diff := strings.TrimSpace(string(output))

	if len(diff) == 0 {
		return "", fmt.Errorf("no staged changes found. Please stage your files using 'git add'")
	}

	return diff, nil
}
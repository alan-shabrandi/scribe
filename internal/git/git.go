package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute git diff command: %w", err)
	}

	rawDiff := strings.TrimSpace(string(output))

	if len(rawDiff) == 0 {
		return "", fmt.Errorf("no staged changes found. Please stage your files using 'git add'")
	}

	filteredDiff, err := FilterDiff(rawDiff)
	if err != nil {
		return "", err
	}

	return filteredDiff, nil
}

func ExecuteCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit failed: %s", string(output))
	}

	return nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

func ExtractTicketID(branchName string) string {
	re := regexp.MustCompile(`(?i)([a-z0-9]+-[0-9]+)`)
	match := re.FindString(branchName)
	return strings.ToUpper(match)
}

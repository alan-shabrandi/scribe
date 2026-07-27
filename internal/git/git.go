package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

var ticketRegex = regexp.MustCompile(`(?i)([a-z0-9]+-[0-9]+)`)

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
		return fmt.Errorf("git commit failed: %w\noutput: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func ExtractTicketID(branchName string) string {
	match := ticketRegex.FindString(branchName)
	return strings.ToUpper(match)
}

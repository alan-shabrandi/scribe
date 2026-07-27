package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func OpenInEditor(initialContent string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}

	if editor == "" {
		editor = "code --wait"
	}

	tmpFile, err := os.CreateTemp("", "SCRIBE_EDIT_MSG_*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}

	tmpFilename := tmpFile.Name()

	defer os.Remove(tmpFilename)

	if _, err := tmpFile.WriteString(initialContent); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temporary file: %w", err)
	}

	tmpFile.Close()

	editorParts := strings.Fields(editor)
	exe := editorParts[0]
	args := append(editorParts[1:], tmpFilename)

	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	updatedBytes, err := os.ReadFile(tmpFilename)
	if err != nil {
		return "", fmt.Errorf("failed to read updated file: %w", err)
	}

	updatedContent := strings.TrimSpace(string(updatedBytes))
	if updatedContent == "" {
		return "", fmt.Errorf("commit message cannot be empty")
	}

	return updatedContent, nil
}

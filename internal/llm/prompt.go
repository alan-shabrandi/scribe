package llm

import (
	"fmt"
	"strings"
)

const (
	conventionalCommitStyle = `1. Follow the Conventional Commits format: <type>(<scope>): <short summary>
   - Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
   - Scope is optional, but encouraged if changes are isolated to a module/package.
2. Keep the header line under 72 characters.`

	freeformStyle = `1. Write a clear, concise, and descriptive commit message in human-readable freeform style.
2. Avoid strict prefixes like feat: or fix: unless natural.
3. Keep the summary clear and under 72 characters.`

	baseStrictRules = `- Write in the imperative mood, present tense (e.g., "add feature", NOT "added feature").
- Do NOT include markdown code blocks.
- Output raw text ONLY.`
)

func getStyleInstructions(style string) string {
	if style == "freeform" {
		return freeformStyle
	}
	return conventionalCommitStyle
}

func getTicketInstruction(ticketID string) string {
	if ticketID == "" {
		return ""
	}
	return fmt.Sprintf("\n- IMPORTANT: Include the Ticket ID [%s] in the commit message header or scope (e.g. feat(%s): ... or fix: [%s] ...).", ticketID, ticketID, ticketID)
}

func BuildSystemPrompt(diff, style, ticketID string) string {
	styleInst := getStyleInstructions(style)
	ticketInst := getTicketInstruction(ticketID)

	return fmt.Sprintf(`You are an automated Git commit message generator.
Your job is to analyze the provided git diff and write a concise, professional commit message.

STRICT RULES:
%s%s
%s

Git Diff:
%s`, styleInst, ticketInst, baseStrictRules, diff)
}

func buildFileSummaryPrompt(fileDiff string) string {
	return fmt.Sprintf(`Summarize the key functional changes in this single file git diff in 1-2 bullet points.
Be concise, accurate, and focus only on the logic modified.

Git Diff for File:
%s`, fileDiff)
}

func BuildSummaryBasedPrompt(summaries []string, style, ticketID string) string {
	styleInst := getStyleInstructions(style)
	ticketInst := getTicketInstruction(ticketID)

	var combinedSummaries string
	if len(summaries) > 0 {
		combinedSummaries = "- " + strings.Join(summaries, "\n- ") + "\n"
	}

	return fmt.Sprintf(`You are an automated Git commit message generator.
Below are summaries of code changes made across multiple files in a repository.
Synthesize these summaries into a single, cohesive, professional commit message.

STRICT RULES:
%s%s
%s

Summarized File Changes:
%s`, styleInst, ticketInst, baseStrictRules, combinedSummaries)
}

func BuildMultiChoicePrompt(diff, style, ticketID string) string {
	styleInst := getStyleInstructions(style)
	ticketInst := getTicketInstruction(ticketID)

	return fmt.Sprintf(`You are an automated Git commit message generator.
Analyze the provided git diff and generate EXACTLY 3 distinct candidate commit messages (e.g., one concise, one detailed, one alternative perspective).

STRICT RULES:
%s%s
- Output EXACTLY 3 lines.
- Each line MUST contain one candidate commit message only.
- Do NOT number the candidates (no "1.", "2.", etc.).
- Do NOT include markdown code blocks or explanations.

Git Diff:
%s`, styleInst, ticketInst, diff)
}

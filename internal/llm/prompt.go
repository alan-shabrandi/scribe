package llm

import "fmt"

func buildSystemPrompt(diff, style string) string {
	var styleInstructions string

	switch style {
	case "freeform":
		styleInstructions = `1. Write a clear, concise, and descriptive commit message in human-readable freeform style.
2. Avoid strict prefixes like feat: or fix: unless natural.
3. Keep the summary clear and under 72 characters.`
	default: // conventional
		styleInstructions = `1. Follow the Conventional Commits format: <type>(<scope>): <short summary>
   - Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
   - Scope is optional, but encouraged if changes are isolated to a module/package.
2. Keep the header line under 72 characters.`
	}

	return fmt.Sprintf(`You are an automated Git commit message generator.
Your job is to analyze the provided git diff and write a concise, professional commit message.

STRICT RULES:
%s
- Write in the imperative mood, present tense (e.g., "add feature", NOT "added feature").
- Do NOT include markdown code blocks (e.g., no triple backticks).
- Do NOT include introductory words, explanations, or concluding remarks.
- Output raw text ONLY.

Git Diff:
%s`, styleInstructions, diff)
}

func buildFileSummaryPrompt(fileDiff string) string {
	return fmt.Sprintf(`Summarize the key functional changes in this single file git diff in 1-2 bullet points.
Be concise, accurate, and focus only on the logic modified.

Git Diff for File:
%s`, fileDiff)
}

func BuildSummaryBasedPrompt(summaries []string, style string) string {
	var styleInstructions string

	switch style {
	case "freeform":
		styleInstructions = `1. Write a clear, concise, and descriptive commit message in human-readable freeform style.
2. Avoid strict prefixes like feat: or fix: unless natural.
3. Keep the summary clear and under 72 characters.`
	default: // conventional
		styleInstructions = `1. Follow the Conventional Commits format: <type>(<scope>): <short summary>
   - Valid types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert.
   - Scope is optional, but encouraged if changes are isolated to a module/package.
2. Keep the header line under 72 characters.`
	}

	combinedSummaries := ""
	for _, s := range summaries {
		combinedSummaries += fmt.Sprintf("- %s\n", s)
	}

	return fmt.Sprintf(`You are an automated Git commit message generator.
Below are summaries of code changes made across multiple files in a repository.
Synthesize these summaries into a single, cohesive, professional commit message.

STRICT RULES:
%s
- Write in the imperative mood, present tense (e.g., "add feature", NOT "added feature").
- Do NOT include markdown code blocks.
- Output raw text ONLY.

Summarized File Changes:
%s`, styleInstructions, combinedSummaries)
}

# Contributing to Scribe

Thank you for your interest in contributing to **Scribe**! 🎉

Scribe is an open-source AI-powered Git commit assistant, and contributions of all kinds are welcome — from bug fixes and new features to documentation improvements and ideas.

Whether you're fixing a small issue or proposing a major improvement, we appreciate your time and effort.

---

## 📋 Table of Contents

- [🚀 Getting Started](#-getting-started)
- [🛠️ Development Guidelines](#️-development-guidelines)
- [🌿 Branching Strategy](#-branching-strategy)
- [💬 Commit Messages](#-commit-messages)
- [📥 Submitting a Pull Request](#-submitting-a-pull-request)
- [🔍 Pull Request Checklist](#-pull-request-checklist)
- [🐛 Reporting Issues](#-reporting-issues)
- [💡 Feature Requests](#-feature-requests)
- [🤝 Code of Conduct](#-code-of-conduct)

---

## 🚀 Getting Started

Before contributing, make sure you have the following installed:

- **Go 1.21+**
- **Git**

### 1. Fork the Repository

Fork the Scribe repository on GitHub:

**Repository:**  
https://github.com/alan-shabrandi/scribe

### 2. Clone Your Fork

Clone your fork locally and enter the project directory:

```bash
git clone https://github.com/YOUR_USERNAME/scribe.git
cd scribe
```

### 3. Build the Project

Build Scribe to make sure your local development environment is working correctly:

```bash
go build ./cmd/scribe
```

### 4. Run the Tests

Before making changes, verify that the existing test suite passes:

```bash
go test ./...
```

If all tests pass, you're ready to start contributing! 🚀

---

## 🛠️ Development Guidelines

Please keep contributions consistent with the existing codebase and Go best practices.

### Formatting

Format your Go code before committing:

```bash
go fmt ./...
```

You can also use:

```bash
gofmt -w .
```

### Testing

All changes should be tested where applicable.

Run the complete test suite with:

```bash
go test ./...
```

When adding new functionality, please include appropriate unit tests whenever possible.

### Build Verification

Make sure the project still builds successfully after your changes:

```bash
go build ./cmd/scribe
```

### Code Quality

When contributing code:

- Keep changes focused and easy to review.
- Prefer simple and idiomatic Go.
- Avoid unnecessary abstractions.
- Follow the existing project structure and conventions.
- Keep functions and packages focused on a clear responsibility.
- Handle errors explicitly and consistently.
- Avoid unrelated changes in the same pull request.
- Update documentation when behavior or configuration changes.

---

## 🌿 Branching Strategy

Please create a dedicated branch for each feature, bug fix, or improvement.

### Feature

```bash
git checkout -b feat/your-feature-name
```

### Bug Fix

```bash
git checkout -b fix/your-bug-name
```

### Documentation

```bash
git checkout -b docs/your-documentation-change
```

### Refactoring

```bash
git checkout -b refactor/your-refactor-name
```

Keeping branches focused makes reviews easier and helps maintain a clean project history.

---

## 💬 Commit Messages

Scribe follows the **Conventional Commits** style.

Use a clear type followed by a concise description of the change.

### Examples

```text
feat: add Gemini provider support
fix: handle empty staged diff
perf: improve diff caching
refactor: simplify provider interface
docs: update installation instructions
test: add cache package tests
chore: update dependencies
```

Common commit types include:

| Type       | Purpose                                 |
| ---------- | --------------------------------------- |
| `feat`     | Introduce a new feature                 |
| `fix`      | Fix a bug                               |
| `perf`     | Improve performance                     |
| `refactor` | Refactor code without changing behavior |
| `docs`     | Documentation changes                   |
| `test`     | Add or update tests                     |
| `chore`    | Maintenance tasks                       |

Keep commit messages:

- Clear and concise
- Written in the imperative mood
- Focused on one logical change
- Consistent with Conventional Commits

---

## 📥 Submitting a Pull Request

Before opening a Pull Request, make sure your changes are ready for review.

### 1. Create a Branch

```bash
git checkout -b feat/your-feature-name
```

### 2. Make Your Changes

Implement your feature, fix, or improvement.

Keep your changes focused and avoid modifying unrelated parts of the codebase.

### 3. Format Your Code

```bash
go fmt ./...
```

### 4. Run Tests

```bash
go test ./...
```

### 5. Verify the Build

```bash
go build ./cmd/scribe
```

### 6. Commit Your Changes

Use a Conventional Commit message:

```bash
git add .
git commit -m "feat: add new provider support"
```

### 7. Push Your Branch

```bash
git push origin feat/your-feature-name
```

### 8. Open a Pull Request

Open a Pull Request against the **`main`** branch of Scribe.

When creating your PR, please provide:

- A clear title
- A concise description of what changed
- The motivation behind the change
- Relevant implementation details
- Testing performed
- Any known limitations or follow-up work

A good Pull Request should make it easy for reviewers to understand **what changed, why it changed, and how it was verified**.

---

## 🔍 Pull Request Checklist

Before submitting your PR, please verify:

- [ ] The project builds successfully.
- [ ] Existing tests pass.
- [ ] New functionality includes tests where appropriate.
- [ ] Go code has been formatted.
- [ ] Commit messages follow Conventional Commits.
- [ ] Documentation has been updated if necessary.
- [ ] The PR contains only relevant changes.
- [ ] No API keys, credentials, or sensitive information are included.
- [ ] The PR description clearly explains the change.

---

## 🐛 Reporting Issues

Found a bug? We'd love to hear about it.

Please open an issue on the **[GitHub Issues](https://github.com/alan-shabrandi/scribe/issues)** page.

When reporting a bug, include as much relevant information as possible.

### A useful bug report should include:

- A clear description of the problem
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Go version
- Operating system
- Scribe version or commit
- Relevant error messages or logs

### Example

```text
## Description

Scribe fails to generate a commit message when the staged diff is empty.

## Steps to Reproduce

1. Run `git add .`
2. Run `scribe generate`
3. Observe the error

## Expected Behavior

Scribe should clearly indicate that there are no staged changes.

## Actual Behavior

The command returns an unexpected error.

## Environment

- OS: Linux
- Go: 1.22
- Scribe: v1.0.0
```

Please avoid including API keys, access tokens, passwords, or other sensitive information in issues.

---

## 💡 Feature Requests

Have an idea that could make Scribe better?

Open an issue on the **[GitHub Issues](https://github.com/alan-shabrandi/scribe/issues)** page and describe:

- The problem you're trying to solve
- The proposed solution
- Why the feature would be useful
- Any alternative approaches you've considered

For larger changes, opening an issue before submitting a Pull Request can help align on the proposed direction first.

---

## 🤝 Code of Conduct

Please be respectful, constructive, and welcoming when participating in the Scribe project.

Contributors are expected to:

- Treat others with respect.
- Provide constructive feedback.
- Keep discussions focused and professional.
- Welcome different perspectives and experience levels.
- Avoid harassment, discrimination, or personal attacks.

Let's keep Scribe a friendly and productive open-source project for everyone. ❤️

---

## ⭐ Thank You!

Every contribution helps make Scribe better.

Whether you're fixing a bug, improving documentation, adding a new LLM provider, improving performance, or simply reporting an issue — **thank you for contributing!**

If you find Scribe useful, consider giving the project a ⭐ on GitHub.

<p align="center">
  <br>
  <a href="https://github.com/alan-shabrandi/scribe">
    <img src="https://img.shields.io/badge/Give_a_⭐_on_GitHub-238636?style=for-the-badge&logo=github" alt="Star on GitHub" />
  </a>
</p>

---

<p align="center">
  Made with ❤️ by the Scribe community
</p>

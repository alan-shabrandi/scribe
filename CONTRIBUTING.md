# Contributing to Scribe

Thank you for your interest in contributing to **Scribe**!

Scribe is an open-source AI-powered Git commit assistant, and
contributions of all kinds are welcome---from bug fixes and new features
to documentation improvements and ideas.

Whether you're fixing a small issue or proposing a major improvement, we
appreciate your time and effort.

---

## Table of Contents

- Getting Started
- Development Guidelines
- Branching Strategy
- Commit Messages
- Submitting a Pull Request
- Pull Request Checklist
- Reporting Issues
- Feature Requests
- Code of Conduct

---

## Getting Started

Before contributing, make sure you have:

- **Go 1.21+**
- **Git**

### 1. Fork the repository

Repository:

https://github.com/alan-shabrandi/scribe

### 2. Clone your fork

```bash
git clone https://github.com/YOUR_USERNAME/scribe.git
cd scribe
```

### 3. Build the project

```bash
go build ./cmd/scribe
```

### 4. Run the tests

```bash
go test ./...
```

If everything passes, you're ready to contribute.

---

## Development Guidelines

- Keep changes focused and easy to review.
- Prefer simple, idiomatic Go.
- Avoid unnecessary abstractions.
- Follow the existing project structure.
- Handle errors explicitly.
- Update documentation when behavior changes.

Format code:

```bash
go fmt ./...
```

or

```bash
gofmt -w .
```

Run tests:

```bash
go test ./...
```

Verify the build:

```bash
go build ./cmd/scribe
```

---

## Branching Strategy

Use a dedicated branch for each change.

```bash
feat/your-feature-name
fix/your-bug-name
docs/your-documentation-change
refactor/your-refactor-name
```

---

## Commit Messages

Scribe follows Conventional Commits.

```text
feat: add Gemini provider support
fix: handle empty staged diff
perf: improve diff caching
refactor: simplify provider interface
docs: update installation instructions
test: add cache package tests
chore: update dependencies
```

Keep commit messages clear, concise, and focused on one logical change.

---

## Submitting a Pull Request

Before opening a pull request:

1.  Format your code.
2.  Run the test suite.
3.  Verify the project builds.
4.  Keep the PR focused.
5.  Explain what changed, why it changed, and how it was tested.

If you're planning a larger change, consider opening an issue first so
the approach can be discussed.

---

## Pull Request Checklist

- [ ] Project builds successfully
- [ ] Tests pass
- [ ] New functionality includes tests where appropriate
- [ ] Go code is formatted
- [ ] Commit messages follow Conventional Commits
- [ ] Documentation updated if needed
- [ ] No sensitive information included

---

## Reporting Issues

Please open an issue with:

- Description
- Steps to reproduce
- Expected behavior
- Actual behavior
- Go version
- Operating system
- Scribe version
- Relevant logs

Avoid including API keys or other sensitive information.

---

## Feature Requests

Feature requests are welcome.

Please describe:

- The problem
- Your proposed solution
- Why it would be be useful
- Alternative approaches considered

---

## Code of Conduct

Be respectful, constructive, and welcoming.

---

## Thank You!

Every contribution helps make Scribe better.

If you find Scribe useful, consider giving the project a ⭐ on GitHub.

```{=html}
<p align="center">
```

Made with ❤️ by the Scribe community

```{=html}
</p>
```

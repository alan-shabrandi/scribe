# Scribe

> AI-powered Git Commit Assistant that generates meaningful commit messages from your staged changes using your preferred LLM.

Generate commit messages from staged Git changes using OpenAI, Claude, Gemini, or Ollama.

<p align="center">
  <img src="docs/demo.gif" alt="Scribe Demo" width="850">
</p>

## Overview

Scribe fits into your existing Git workflow in different ways:

- **CLI** — Generate commit messages from any terminal.
- **Git Hook** — Integrate with `git commit` and generate messages automatically.
- **VS Code Extension** — Generate commit messages directly from the Source Control panel.

The CLI is the core of the project, while the VS Code extension provides a native experience for VS Code users.

👉 https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode

## Table of Contents

- Features
- Installation
- Quick Start
- Configuration
- Usage
- Git Hook Integration
- Performance
- Project Structure
- Contributing
- License

## Features

- Supports OpenAI, Claude, Gemini, and Ollama
- Interactive CLI built with Cobra and Survey
- Detects ticket IDs from branch names
- SHA-256 caching for identical staged diffs
- Handles large diffs through chunking and summarization
- Can be used as both a CLI and Git hook
- Supports `.scribeignore` to skip custom files or patterns from diff analysis

## Installation

### Pre-built binaries

Download the latest release from GitHub Releases.

### Go

```bash
go install github.com/alan-shabrandi/scribe/cmd/scribe@latest
```

### Build from source

```bash
git clone https://github.com/alan-shabrandi/scribe.git
cd scribe
go build -o scribe ./cmd/scribe
```

## Quick Start

```bash
scribe config set provider openai
scribe config set api_key YOUR_API_KEY

git add .
scribe generate
```

Scribe analyzes your staged changes and suggests commit message candidates.

## Configuration

Configuration is stored in:

```text
~/.scribe.yaml
```

Common settings:

- provider
- model
- style
- api_key

## Usage

```bash
git add .
scribe generate
```

## Git Hook Integration

```bash
scribe hook install
```

Remove the hook:

```bash
scribe hook uninstall
```

## Ignoring Files

You can create a `.scribeignore` file in the root of your repository to ignore specific files or patterns from being processed by Scribe (similar to `.gitignore`):

```text
# Ignore lockfiles and generated files
*.lock
docs/*.md
vendor/
```

## Performance

Scribe caches responses for identical staged diffs using SHA-256, reducing unnecessary LLM requests.

Cache location:

```text
~/.scribe_cache.json
```

## Project Structure

```text
scribe/
├── cmd/
├── internal/
├── go.mod
└── README.md
```

## Contributing

Issues and pull requests are welcome.

If Scribe is useful to you, consider giving the repository a star.

## License

Released under the MIT License.

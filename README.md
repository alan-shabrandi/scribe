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
- **Auto-Copy Setting** - Commit messages can be auto copied to clipboard to paste manually as you choose.

The CLI is the core of the project, while the VS Code extension provides a native experience for VS Code users.

👉 https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode

## Table of Contents

- Features
- Installation
- Quick Start
- Configuration
- API Key Storage
- Usage
- Git Hook Integration
- Performance
- Project Structure
- Contributing
- License

## Features

- Supports OpenAI, Claude, Gemini, and Ollama
- Stores API keys in your OS credential store, not in a plaintext file
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

The API key goes into your operating system's credential store rather than into
`~/.scribe.yaml` — see [API Key Storage](#api-key-storage).

Scribe analyzes your staged changes and suggests commit message candidates.


```bash
scribe generate -c
```
Generates a commit message, once completed auto copies to the users clipboard.

## Configuration

Configuration is stored in:

```text
~/.scribe.yaml
```

Common settings:

- provider
- model
- style
- auto_copy

`api_key` is deliberately absent from that list: it is stored outside the config
file. See [API Key Storage](#api-key-storage).

Inspect the current configuration (keys are always masked):

```bash
scribe config show
```

## API Key Storage

Scribe keeps API keys in your operating system's native credential store
instead of in plaintext:

| Platform | Store                                        |
| -------- | -------------------------------------------- |
| macOS    | Keychain                                     |
| Windows  | Credential Manager                           |
| Linux    | Secret Service (GNOME Keyring, KWallet, ...) |

`scribe config set api_key` and `scribe init` write there by default. Keys are
stored per provider under the entry name `scribe`, so keys for OpenAI, Claude
and Gemini can coexist and switching `provider` picks up the matching key.

```bash
# Store a key for the configured provider
scribe config set api_key YOUR_API_KEY

# Store a key for another provider without switching to it
scribe config set api_key YOUR_API_KEY --provider claude

# See what is stored, and which key is actually in use
scribe config keyring status

# Move an existing plaintext api_key out of ~/.scribe.yaml
scribe config keyring migrate

# Remove a stored key
scribe config keyring delete openai
```

### Where Scribe looks for a key

Highest precedence first:

1. `SCRIBE_API_KEY` — a deliberate per-shell override.
2. `api_key` in `~/.scribe.yaml` — plaintext, but explicit, so editing the file
   always has the effect you expect.
3. The OS credential store, keyed by provider.
4. The provider's own variable: `OPENAI_API_KEY` or `GEMINI_API_KEY`.

Because the config file outranks the credential store, `scribe config set
api_key` removes a plaintext `api_key` from `~/.scribe.yaml` after storing the
new key, so the old one cannot keep being used silently.

### When the credential store is not available

Headless Linux machines, containers and CI runners often have no Secret Service
running. Scribe does not fail there: it falls through to the environment
variables above, and `scribe config keyring status` explains what it found. To
write a key to the config file anyway:

```bash
scribe config set api_key YOUR_API_KEY --plaintext
```

That file is written with `0600` permissions, but it is still plaintext —
anything able to read it can read the key.

## Usage

```bash
git add .
scribe generate

# Display help and available commands
scribe --help
scribe config --help
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

# Scribe

> _AI-powered Git Commit Assistant that generates meaningful commit messages from your staged changes using your preferred LLM._

Generate commit messages from staged Git changes using OpenAI, Claude, Gemini, or Ollama.

<p align="center">
  <img src="docs/demo.gif" alt="Scribe Demo" width="850">
</p>

## Overview

Scribe fits into your existing Git workflow in different ways:

- **CLI** — Generate commit messages from any terminal.
- **Git Hook** — Integrate with `git commit` and generate messages automatically.
- **VS Code Extension** — Generate commit messages directly from the Source Control panel.
- **Auto-Copy Setting** — Commit messages can be copied automatically to the clipboard so you can paste them manually.

The CLI is the core of the project, while the VS Code extension provides a native experience for VS Code users.

👉 [Scribe on the VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Secret Scanning](#secret-scanning)
- [API Key Storage](#api-key-storage)
- [Usage](#usage)
- [Git Hook Integration](#git-hook-integration)
- [Ignoring Files](#ignoring-files)
- [Performance](#performance)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [License](#license)

## Features

- Supports OpenAI, Claude, Gemini, and Ollama.
- **Secret Scanner:** Detects accidentally staged API keys or credentials and blocks sending them to the LLM.
- **OS Keyring Integration:** Stores API keys securely in your OS credential store (macOS Keychain, Windows Credential Manager, Linux Secret Service) instead of plaintext files.
- Interactive CLI built with Cobra and Survey.
- Detects ticket IDs from branch names.
- SHA-256 caching for identical staged diffs.
- Handles large diffs through chunking and summarization.
- Can be used as both a CLI and Git hook.
- Supports `.scribeignore` to skip custom files or patterns from diff analysis.

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

The API key goes into your operating system's credential store rather than into `~/.scribe.yaml` — see [API Key Storage](#api-key-storage).

Scribe analyzes your staged changes and suggests commit message candidates.

```bash
scribe generate -c
```

Generates a commit message and, once completed, automatically copies it to the user's clipboard.

## Configuration

Configuration is stored in:

```text
~/.scribe.yaml
```

Common settings:

- `provider`
- `model`
- `style`
- `auto_copy`
- `secret_scan` (default `true`; see [Secret Scanning](#secret-scanning))

`api_key` is deliberately absent from that list: it is stored outside the config file. See [API Key Storage](#api-key-storage).

## Secret Scanning

Before `scribe generate` sends your staged diff to an LLM provider, it scans the diff locally for credentials. If it finds any, it stops, sends nothing, and lists each offending line with its file, line number, and matching rule. The secret itself is masked in that output.

Detected:

- AWS access key IDs and secret access keys.
- GitHub tokens (classic and fine-grained), Slack tokens, and webhook URLs.
- Generic API keys and tokens assigned to names like `api_key` or `client_secret`.
- PEM private keys (RSA, DSA, EC, OpenSSH, PGP).
- Database connection strings with an embedded password (Postgres, MySQL, MongoDB, Redis, AMQP, SQL Server).

Obvious placeholders such as `${DB_PASSWORD}`, `<your-key>`, or `xxxx` are ignored.

If a finding is a false positive, skip the scan for one run:

```bash
scribe generate --skip-secret-scan
```

Or turn it off entirely:

```bash
scribe config set secret_scan false
```

Inspect the current configuration (keys are always masked):

```bash
scribe config show
```

## API Key Storage

Scribe keeps API keys in your operating system's native credential store instead of in plaintext:

| Platform | Store                                        |
| -------- | -------------------------------------------- |
| macOS    | Keychain                                     |
| Windows  | Credential Manager                           |
| Linux    | Secret Service (GNOME Keyring, KWallet, ...) |

`scribe config set api_key` and `scribe init` write there by default. Keys are stored per provider under the entry name `scribe`, so keys for OpenAI, Claude, and Gemini can coexist, and switching `provider` picks up the matching key.

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
2. `api_key` in `~/.scribe.yaml` — plaintext, but explicit, so editing the file always has the effect you expect.
3. The OS credential store, keyed by provider.
4. The provider's own variable: `OPENAI_API_KEY` or `GEMINI_API_KEY`.

Because the config file outranks the credential store, `scribe config set api_key` removes a plaintext `api_key` from `~/.scribe.yaml` after storing the new key, so the old one cannot keep being used silently.

### When the credential store is not available

Headless Linux machines, containers, and CI runners often have no Secret Service running. Scribe does not fail there: it falls through to the environment variables above, and `scribe config keyring status` explains what it found. To write a key to the config file anyway:

```bash
scribe config set api_key YOUR_API_KEY --plaintext
```

That file is written with `0600` permissions, but it is still plaintext — anything able to read it can read the key.

## Usage

### CLI

```bash
git add .
scribe generate

# Display help and available commands
scribe --help
scribe config --help
```

### VS Code Extension

Once the extension is installed, you can:

- Press `Ctrl+Alt+C` (`Cmd+Alt+C` on macOS).
- Click the Scribe button in the Source Control panel.
- Open the Command Palette (`Ctrl+Shift+P` / `Cmd+Shift+P`) and run `Scribe: Generate Commit Message`.

## Git Hook Integration

Install the hook:

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
├── cmd/                 # CLI entry points
├── internal/            # Core logic
│   ├── config/           # Configuration management
│   ├── git/              # Git operations
│   ├── llm/              # LLM provider integrations
│   ├── secrets/          # OS Keyring integration
│   └── secretscan/       # Local secret scanning logic
├── vscode-extension/    # VS Code extension source
├── go.mod
└── README.md
```

## Contributing

Issues and pull requests are welcome.

If Scribe is useful to you, consider giving the repository a star.

## License

Released under the MIT License.

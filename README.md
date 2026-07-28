# ✍️ Scribe

> AI-powered Git Commit Assistant that generates clean, meaningful commit messages from your staged changes using LLMs.

Generate meaningful Git commit messages instantly from your staged changes using OpenAI, Claude, Gemini, or Ollama.

<p align="center">
  <img src="docs/demo.gif" alt="Scribe Demo" width="850">
</p>

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## 📚 Table of Contents

- Features
- Quick Start
- Installation
- Configuration
- Usage
- Git Hook Integration
- Performance & Caching
- Project Structure
- License

---

## ✨ Features

- 🤖 **Multi-Provider Support:** OpenAI, Anthropic Claude, Google Gemini, and Ollama.
- 💬 **Interactive CLI:** Built with Cobra and Survey.
- 🏷️ **Context Awareness:** Detects ticket IDs from branch names.
- ⚡ **SHA-256 Caching:** Reuses responses for identical staged diffs.
- 🧩 **Large Diff Handling:** Chunking and summarization for large changes.
- 🔗 **Git Hook Support:** Works as both a CLI and `prepare-commit-msg` hook.

---

## 🚀 Quick Start

```bash
brew install alan-shabrandi/tap/scribe

scribe config set provider openai
scribe config set api_key YOUR_API_KEY

git add .
scribe generate
```

---

## 📦 Installation

### Homebrew (macOS / Linux)

```bash
brew install alan-shabrandi/tap/scribe
```

### Build from Source

#### Prerequisites

- Go 1.21+
- Git

```bash
git clone https://github.com/alan-shabrandi/scribe.git
cd scribe
go build -o scribe ./cmd/scribe
sudo mv scribe /usr/local/bin/
```

---

## ⚙️ Configuration

Configuration file:

```text
~/.scribe.yaml
```

| Setting  | Example      |
| -------- | ------------ |
| provider | openai       |
| model    | gpt-4o       |
| style    | conventional |
| api_key  | your-api-key |

```bash
scribe config set provider openai
scribe config set api_key "your-api-key"
scribe config set model "gpt-4o"
scribe config set style "conventional"
```

---

## 🚀 Usage

1. Stage your changes

```bash
git add .
```

2. Generate commit messages

```bash
scribe generate
```

Example:

```text
🔍 Fetching staged git changes...
📌 Detected Context: Branch 'feature/PROJ-892-add-cache'
✔ Generating candidate commit messages via 'openai'...

? Select a commit message option:

▸ feat(cache): add SHA-256 caching for staged diffs (PROJ-892)
  perf(diff): skip LLM invocation on identical git diff (PROJ-892)
  ✏️ Edit custom message in system editor
  🚫 Cancel
```

---

## 🔗 Git Hook Integration

```bash
scribe hook install
scribe hook uninstall
```

---

## ⚡ Performance & Caching

Cached responses are stored in:

```text
~/.scribe_cache.json
```

---

## 🛠️ Project Structure

```text
scribe/
├── cmd/
│   └── scribe/
├── internal/
│   ├── cache/
│   ├── config/
│   ├── git/
│   └── llm/
├── go.mod
└── README.md
```

---

## 📄 License

Licensed under the MIT License.

See `LICENSE` for details.

If you find Scribe useful, consider giving the repository a ⭐.

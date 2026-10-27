# Scribe — AI Commit Message Generator

<p align="center">
  <strong>Generate clean, meaningful Git commit messages directly inside VS Code using AI.</strong>
</p>

<p align="center">
  <a href="https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode">
    <img src="https://img.shields.io/visual-studio-marketplace/v/alan-shabrandi.scribe-vscode?style=flat-square&color=blue&label=VS%20Code%20Marketplace" alt="VS Code Marketplace version">
  </a>
  <a href="https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode">
    <img src="https://img.shields.io/visual-studio-marketplace/i/alan-shabrandi.scribe-vscode?style=flat-square&color=green" alt="VS Code Marketplace installs">
  </a>
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="MIT License">
</p>

Scribe generates conventional commit messages from your staged Git changes and inserts them directly into the VS Code Source Control view.

## Demo

<p align="center">
  <img src="https://i.imgur.com/Ra357nT.gif" alt="Scribe demo" width="850">
</p>

## Features

| Feature              | Description                                                                 |
| -------------------- | --------------------------------------------------------------------------- |
| One-click generation | Generate a commit message from staged changes.                              |
| Keyboard shortcut    | Trigger generation with `Ctrl+Alt+C` / `Cmd+Alt+C`.                         |
| Custom ignores       | Respects `.scribeignore` files to skip lockfiles and custom paths.          |
| Multiple providers   | Supports OpenAI, Claude, Gemini, and Ollama.                                |
| Ticket detection     | Detects issue IDs from branch names.                                        |
| Smart caching        | Uses a SHA-256 cache to prevent duplicate API calls.                        |
| Conventional Commits | Generates messages following the Conventional Commits format.               |
| Native integration   | Inserts the generated message directly into the VS Code commit box.         |
| **Secret scanner**   | **Blocks generation if staged changes contain API keys or credentials.**    |
| **Secure keyring**   | **Relies on the Scribe CLI to store API keys in your OS credential store.** |

## Installation

Scribe is a VS Code extension that uses the [Scribe CLI](https://github.com/alan-shabrandi/scribe) to generate commit messages.

### 1. Install the Scribe CLI

Download the latest binary for your operating system from [Scribe CLI releases](https://github.com/alan-shabrandi/scribe/releases).

On Windows, download `scribe.exe` and add its directory to your system `PATH`.

On macOS or Linux, move the downloaded binary into a directory on your `PATH`:

```bash
sudo mv scribe /usr/local/bin/
```

For other installation options, including building from source, see the [Scribe CLI repository](https://github.com/alan-shabrandi/scribe).

### 2. Install the VS Code extension

Install Scribe from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode).

### 3. Configure Scribe

Set your preferred provider and API key:

```bash
scribe config set provider openai
scribe config set api_key YOUR_API_KEY
```

The API key is stored in your operating system's credential store (macOS Keychain, Windows Credential Manager, or the Linux Secret Service), not in `~/.scribe.yaml`. Run `scribe config keyring status` to check what is stored.

## Usage

1. Stage your Git changes:

   ```bash
   git add .
   ```

2. Trigger Scribe using one of these methods:
   - Press `Ctrl+Alt+C` (`Cmd+Alt+C` on macOS).
   - Click the Scribe button in the Source Control panel.
   - Open the Command Palette (`Ctrl+Shift+P` / `Cmd+Shift+P`) and run:

     ```text
     Scribe: Generate Commit Message
     ```

3. Select a suggestion.
4. Scribe inserts the generated message into the VS Code commit box.

> **Note:** Scribe includes a built-in Secret Scanner. If you accidentally stage a file containing an API key or password, the extension blocks the request and alerts you, helping prevent accidental data leaks to the LLM.

## Configuration

You can configure the provider, model, and commit message style through the CLI.

For example:

```bash
scribe config set provider openai
scribe config set model gpt-4o
scribe config set style conventional
```

## Supported Providers

| Provider | Supported |
| -------- | :-------: |
| OpenAI   |    Yes    |
| Claude   |    Yes    |
| Gemini   |    Yes    |
| Ollama   |    Yes    |

## Roadmap

- [x] VS Code extension
- [x] Multi-provider support
- [x] Smart caching
- [ ] Custom prompt templates
- [ ] Azure OpenAI
- [ ] JetBrains plugin

## Support

If you find Scribe useful, consider leaving a review on the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode) or starring the [GitHub repository](https://github.com/alan-shabrandi/scribe).

## License

Released under the MIT License.

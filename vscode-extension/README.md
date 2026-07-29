# ✍️ Scribe — AI Commit Message Generator

<p align="center">
  <b>Generate clean, meaningful Git commit messages directly inside VS Code using AI.</b>
</p>

<p align="center">

[![VS Code Marketplace Version](https://img.shields.io/visual-studio-marketplace/v/alan-shabrandi.scribe-vscode?style=flat-square&color=blue&label=VS%20Code%20Marketplace)](https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode)
[![Installs](https://img.shields.io/visual-studio-marketplace/i/alan-shabrandi.scribe-vscode?style=flat-square&color=green)](https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

</p>
Seamlessly integrate **Scribe** into your VS Code workflow to generate conventional commit messages from staged changes with a single click.

---

## 📸 Demo

<p align="center">
  <img src="https://i.imgur.com/MXJqu7x.gif" alt="Scribe Demo" width="850">
</p>

---

## ✨ Features

| Feature                 | Description                                 |
| ----------------------- | ------------------------------------------- |
| ⚡ One-Click Generation | Generate commit messages instantly          |
| 🤖 Multi-Provider       | OpenAI, Claude, Gemini and Ollama           |
| 🏷️ Ticket Detection     | Detect issue IDs from branch names          |
| 💾 Smart Caching        | SHA-256 cache avoids duplicate API calls    |
| 📝 Conventional Commits | Clean and consistent commit messages        |
| 🔗 Native Integration   | Writes directly into the VS Code commit box |

---

## 🚀 Quick Setup & Installation

Scribe VS Code extension acts as a frontend interface for the **Scribe CLI**.

### 1️⃣ Install Scribe CLI

Download the latest executable binary for your operating system:

- 🪟 **Windows:** Download `scribe.exe` from **[Releases](https://github.com/alan-shabrandi/scribe/releases)** and add its directory to your system `PATH`.
- 🍎 **macOS / 🐧 Linux:** Download the binary and move it to your path:
  ```bash
  sudo mv scribe /usr/local/bin/
  ```

> 💡 _Prefer installing via Go or building from source? Check the [Scribe CLI Repository](https://github.com/alan-shabrandi/scribe) for full details._

### 2️⃣ Initial Configuration

Open your terminal and set your preferred AI provider and API key:

```bash
scribe config set provider openai
scribe config set api_key YOUR_API_KEY
```

---

## 💡 Usage

1. Stage your Git changes (`git add .`).
2. Click the Scribe button in the Source Control panel or press `Ctrl+Shift+P` (`Cmd+Shift+P` on Mac) and run:

```text
Scribe: Generate Commit Message
```

3. Select a suggestion from the menu.
4. The generated message will be automatically inserted into your VS Code commit box!

---

## ⚙️ Advanced Configuration

Customize model selection, styles, and providers using the CLI:

```bash
scribe config set provider openai
scribe config set model gpt-4o
scribe config set style conventional
```

---

## 🤖 Supported Providers

| Provider | Supported |
| -------- | --------- |
| OpenAI   | ✅        |
| Claude   | ✅        |
| Gemini   | ✅        |
| Ollama   | ✅        |

---

## 🛣️ Roadmap

- [x] VS Code Extension
- [x] Multi-provider support
- [x] Smart caching
- [ ] Custom prompt templates
- [ ] Azure OpenAI
- [ ] JetBrains plugin

---

## ⭐️ Support & Feedback

If you find **Scribe** useful, please leave a 5-star review on the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=alan-shabrandi.scribe-vscode) or star the repository on [GitHub](https://github.com/alan-shabrandi/scribe)!

## 📄 License

Released under the MIT License.

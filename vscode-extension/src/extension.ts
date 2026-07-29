import * as vscode from "vscode";
import { execFile } from "child_process";
import * as path from "path";

export function activate(context: vscode.ExtensionContext) {
  let disposable = vscode.commands.registerCommand(
    "scribe.generate",
    async () => {
      const workspaceFolders = vscode.workspace.workspaceFolders;
      if (!workspaceFolders) {
        vscode.window.showErrorMessage(
          "Please open a workspace in VS Code first.",
        );
        return;
      }
      const cwd = workspaceFolders[0].uri.fsPath;

      const scribeCommand = "F:\\Learn\\IT\\Go\\projects\\scribe\\scribe.exe";

      vscode.window.withProgress(
        {
          location: vscode.ProgressLocation.Notification,
          title:
            "Scribe is analyzing changes and generating commit messages...",
          cancellable: false,
        },
        async (progress) => {
          return new Promise<void>((resolve, reject) => {
            execFile(
              scribeCommand,
              ["generate", "--json"],
              { cwd },
              async (error, stdout, stderr) => {
                if (error) {
                  vscode.window.showErrorMessage(
                    `Failed to run Scribe: ${stderr || error.message}`,
                  );
                  resolve();
                  return;
                }

                try {
                  const candidates: string[] = JSON.parse(stdout.trim());

                  if (!candidates || candidates.length === 0) {
                    vscode.window.showWarningMessage(
                      "No commit candidates generated or no staged changes detected.",
                    );
                    resolve();
                    return;
                  }

                  const selected = await vscode.window.showQuickPick(
                    candidates,
                    {
                      placeHolder: "Select a commit message option:",
                    },
                  );

                  if (selected) {
                    const gitExtension =
                      vscode.extensions.getExtension("vscode.git")?.exports;
                    if (gitExtension) {
                      const api = gitExtension.getAPI(1);
                      const repo = api.repositories[0];
                      if (repo) {
                        repo.inputBox.value = selected;
                        vscode.window.showInformationMessage(
                          "Commit message populated into Git input box!",
                        );
                      } else {
                        vscode.window.showWarningMessage(
                          "Git repository not found.",
                        );
                      }
                    }
                  }
                } catch (e: any) {
                  vscode.window.showErrorMessage(
                    `Error parsing output: ${e.message}`,
                  );
                }
                resolve();
              },
            );
          });
        },
      );
    },
  );

  context.subscriptions.push(disposable);
}

export function deactivate() {}

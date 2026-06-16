# User Prompt (общий для Claude, Agent, OpenCode, Codex, Gemini, Qwen)

## Режим: первый запуск (не continue)

```
Issue to solve: {issueUrl}
Your prepared branch: {branchName}
Your prepared working directory: {tempDir}
```

Если включён `--enable-workspaces`:
```
Your prepared tmp directory for logs and downloads: {workspaceTmpDir}
```

Если PR уже существует:
```
Your prepared Pull Request: {prUrl}
```

Если fork mode (`--fork`):
```
Your forked repository: {forkedRepo}
Original repository (upstream): {owner}/{repo}
GitHub Actions on your fork: {forkActionsUrl}
```

Если есть contributing guidelines:
```
{contributingGuidelines}
```

Если есть thinking prompt (`--think`):
```
{Think. / Think hard. / Think harder. / Ultrathink.}
```

Финал:
```
Proceed.
```

---

## Режим: continue

```
Issue to solve: https://github.com/{owner}/{repo}/issues/{issueNumber}
```

или если PR:
```
Issue to solve: Issue linked to PR #{prNumber}
```

Всё остальное то же, но финал:
```
Continue.
```

Если в continue mode есть feedback из PR-комментариев, каждый feedback line
вставляется как отдельная строка перед финалом.

---

## Режим: minimal restart (`--minimal-restart-context`)

При `--resume` и `--minimal-restart-context` user prompt заменяется на:

```
🔄 Auto-restart: resume the previous session and handle its uncommitted changes.

Uncommitted files ({fileCount}):
{uncommittedFiles}

Changes summary:
{diffSummary}

Please review these changes and commit them with an appropriate commit message.
Follow the repository's commit message conventions from previous commits.
```

Либо, если resume не удался:

```
Continuing work on issue: {issueUrl}

Previous session completed but left uncommitted changes.

Feedback from reviewers:
{feedbackLines}

Uncommitted changes:
{uncommittedFiles}

Full diff:
{fullDiff}

Please review these changes and commit them appropriately.
```

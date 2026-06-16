# HANDOFF.md — Agent Skill (`handoff.prompts.lib.mjs` / `handoff-skill.lib.mjs`)

Деплоится как SKILL.md в `.claude/skills/handoff/SKILL.md` (и симлинк в `.agents/skills/handoff/`).
Включается флагом `--use-handoff`.

```
---
name: handoff
description: Maintain a HANDOFF.md continuity document in the repository root so
  any session can continue a previous session's work — even across different AI
  tools (Claude and Codex) in the same pull request. Use when starting, resuming,
  or finishing work on a long-running task, issue, or pull request.
---

# HANDOFF.md continuity skill

HANDOFF.md is a single shared handoff document in the repository root that lets
any session continue the work of any previous session, even when a different AI
tool (for example Claude and Codex) is used. It travels with the pull request
branch, so it is the cross-tool, cross-session memory for this PR.

## When to use this skill

- When you start a working session, read HANDOFF.md first if it exists. Treat
  its "Next steps" section as your immediate starting point and honor the
  decisions and constraints it records before exploring anything else.
- When HANDOFF.md does not exist yet and the task is non-trivial, create it
  early so an interrupted session can always be resumed.
- When you make meaningful progress, update HANDOFF.md so it always reflects
  the current truth. Keep exactly one active HANDOFF.md per pull request branch
  (do not create per-session copies).
- When all requirements are fully met and the work is complete, record that
  completion at the top of HANDOFF.md (or delete the file) so the next session
  knows there is nothing left to continue.

## How to write HANDOFF.md

- Keep it concise and tool-agnostic: describe state by referencing file paths,
  function names, branch, and commit SHAs rather than tool-specific commands,
  so the next tool (Claude or Codex) can act on it directly. Prefer pointers
  to existing artifacts over duplicating their content.
- Include these sections:
  1. **Task** — the issue/PR being solved and the goal.
  2. **Current state** — what is done and verified.
  3. **Decisions** — key choices made and why (so they are not re-litigated).
  4. **Next steps** — the concrete, ordered actions the next session should take.
  5. **Gotchas** — known pitfalls, failing checks, or constraints.
  6. **Critical files** — the important paths and what each is for.
- When you record next steps, make them specific and actionable (a path, a
  function, a command to run) instead of vague goals, and remove items as they
  are completed.

## Committing and safety

- When you finish a step that changes the state, commit HANDOFF.md together
  with the related code changes so the handoff stays in sync with the branch
  and is never lost if the session is interrupted.
- Never include secrets, tokens, API keys, passwords, or personal data in
  HANDOFF.md — it is committed to the repository.
```

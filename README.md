# Octra — AI agent factory powered by Nix

Octra is a multi-agent AI orchestrator that builds software projects using a Boss → Manager → Worker pipeline. Every project is snapshotted into the [Nix](https://nixos.org) store, enabling instant rollback and zero-storage recovery.

## Architecture (Boss → Manager → Worker)

![Octra logic](docs/icons/octra-logic.png)

```
User → API Gateway → Boss (architect)
                        ├── Manager (backend) → Worker × N
                        ├── Manager (frontend) → Worker × N
                        └── Manager (devops)  → Worker × N
                              ↓
                        Agents Service → Claude / Gemini / GPT / ...
```

**Boss** plans architecture, spawns managers, validates output, pushes to GitHub.  
**Managers** review and orchestrate workers.  
**Workers** generate code inside Nix-isolated environments — either via AI or real tool scaffolding.  

## Tool scaffolding pipeline

Octra can scaffold real projects using native toolchains inside `nix develop`:

```
setupProject() → FlakeBuilder(techStack) → ToolExecutor(generateViaTools)
                                              ├── npm init / cargo init / flutter create / …
                                              ├── AI generation fallback if Nix/tool unavailable
                                              └── git status --porcelain to detect created files
```

- **FlakeBuilder** generates `flake.nix` with Nix packages for 20+ tech stacks
- **ToolExecutor** runs real scaffolding commands (`npm install`, `cargo init`, `composer create-project`, etc.)
- Graceful fallback to AI generation if Nix or tool is unavailable

## Structured tool guides (`pkg/guids`)

Octra ships with a registry of structured tool guides — one file per tool, organized by language ecosystem:

```
pkg/guids/
├── core/               # Guide type, registry, formatting
├── golang/             # go
├── rust/               # cargo
├── node/               # npm
├── python/             # pip
├── java/               # maven, gradle, kotlin/
├── php/                # composer, artisan
├── flutter/            # flutter
├── cpp/                # cmake
├── dotnet/             # dotnet
├── elixir/             # mix
├── haskell/            # cabal
├── zig/                # zig
├── swift/              # swiftpm
└── ruby/               # bundler, gem
```

Each guide contains structured commands, Nix packages, and project structure — injected into AI prompts to prevent hallucinated commands and save tokens.

## Context management system (`internal/service/context/`)

Octra maintains per-project context with three visibility scopes, stored in Redis (fast cache) and PostgreSQL (permanent storage):

```
                        ┌──────────────────┐
                        │  AI Agent returns │
                        │  JSON with        │
                        │  "context" field  │
                        └──────┬───────────┘
                               ↓
               ┌───────────────────────────────┐
               │  ContextService.SaveFromAI()   │
               │  ┌──────────┐  ┌────────────┐ │
               │  │ Postgres │  │   Redis    │ │
               │  │ (always) │  │ (5min TTL) │ │
               │  └──────────┘  └────────────┘ │
               └───────────────────────────────┘
                               ↓
               ┌───────────────────────────────┐
               │  GetForPrompt() → injects     │
               │  into Boss/Manager/Worker      │
               └───────────────────────────────┘
```

**Three scope levels:**

| Scope | Visibility | Set by | Lifecycle |
|-------|-----------|--------|-----------|
| `global` | All agents in project | Boss/Manager | Persists forever |
| `team` | All workers under one manager | Manager | Until manager finishes |
| `individual` | Single agent | Any agent | Per agent session |

**How context flows:**

1. **Boss** sends prompt → AI returns JSON with optional `"context": {"scope": "global", "type": "global_rule", "content": "..."}`
2. **ContextService** saves to Postgres + Redis, then injects into all subsequent agent prompts
3. **Manager** and **Worker** also receive and can create context (team/individual scopes)
4. Before every AI call, `GetForPrompt()` assembles all relevant context into a formatted block

**Automatic cleanup:**

- **Messages** capped at ~20 (adaptive: 40 if avg length <200 chars, 30 if <500 chars)
- When limit exceeded, the **5 oldest non-important** messages are removed (sawtooth pattern)
- **Global rules** live forever unless user says `"forget"`
- If user/AI sends `"content": "forget"` → all matching entries are marked `forgotten`
- Redis entries expire after 5 minutes, reloaded from Postgres on next access

**Key files:**

| File | Purpose |
|------|---------|
| `internal/service/context/response.go` | AI context flag parsing, forget detection |
| `internal/service/context/postgres.go` | Permanent storage, adaptive cleanup |
| `internal/service/context/redis.go` | TTL cache (5 min), cache-aside pattern |
| `internal/service/context/service.go` | Coordinator, SaveFromAI, GetForPrompt |
| `pkg/models/context_entry.go` | GORM model (project_id, scope, type, content, …) |

## Group Chat Agent Orchestration (`internal/service/groupchat/`)

Octra's workers use the **Group Chat** pattern ([Microsoft Agent Framework](https://learn.microsoft.com/en-us/agent-framework/workflows/orchestrations/group-chat)) for true multi-agent collaboration:

```
Round 1: all agents generate concurrently (AllAtOnceSelector)
         Worker A ──→ Message{files}
         Worker B ──→ Message{files}
         Worker C ──→ Message{files}
                ↓ broadcast
Round 2: each agent sees full conversation → can refine based on peer work
         Worker A ──→ sees B+C files → improves
         Worker B ──→ sees A+C files → improves
         Worker C ──→ sees A+B files → improves
                ↓ termination condition
         git add . + git commit
```

**How agents "check the chat":**

Every agent's `Process(ctx, conv)` receives the full `Conversation` with all messages and files. `WorkerAgent.buildChatContext()` extracts peer-generated files and content, injects them into the AI prompt — workers literally see what others built and can adapt.

**Key components:**

| File | Purpose |
|------|---------|
| `internal/service/groupchat/types.go` | Agent, Message, Conversation, SpeakerSelector interfaces |
| `internal/service/groupchat/orchestrator.go` | Orcherator: registration, concurrent rounds, termination |
| `internal/service/groupchat/selection.go` | RoundRobin, AllAtOnce, RoleBased, AgentBased selectors |
| `internal/service/rules/worker/agent.go` | `WorkerAgent` implementing `groupchat.Agent` with AI generation |

## Project lifecycle with Nix

```
setupProject() → generate flake.nix → AI generates code → nix-store --add → cleanup
                                                                 ↓
                                                          RestoreProject(taskID)
```

Each project:

1. Gets a `flake.nix` at creation time (auto-generated per tech stack)
2. Is built by AI agents inside the working directory
3. On completion: **snapshotted to `/nix/store/`** via `nix-store --add`
4. Working directory is removed (zero disk waste)
5. **`RestoreProject(taskID)`** recovers the project from the Nix store instantly

This means projects take zero space when idle but can be restored at any time.

## Services

| Service | Stack | Role |
|---------|-------|------|
| `orchestrator` | Go, gRPC | Boss → Manager → Worker pipeline, Group Chat orchestration, Nix snapshots, tool scaffolding |
| `agents` | Go, gRPC | AI provider proxy (Claude, Gemini, GPT, DeepSeek, …) |
| `apigateway` | Go, Gin, WebSocket | HTTP/WS → gRPC bridge |
| `user` | Go, Gin | Auth, subscriptions, custom providers |
| `frontend/web` | React, Vite, Electron | Interactive canvas + chat UI |
| `grademodel` | Python, scikit-learn | Task complexity grading |

## Nix integration

| File | Purpose |
|------|---------|
| `orchestrator/flake.nix` | Builds orchestrator binary via `buildGoModule` |
| `orchestrator/nix/module.nix` | NixOS module (systemd service) |
| `orchestrator/Dockerfile` | Based on `nixos/nix` — includes Nix in container |
| `orchestrator/internal/service/rules/boss/project.go` | `snapshotProject()`, `RestoreProject()`, `generateFlake()` |
| `orchestrator/internal/service/rules/boss/flake_builder.go` | Dynamic `flake.nix` generation per tech stack |
| `orchestrator/internal/service/rules/boss/nix_build.go` | `nix build`, `nix flake check`, `nix flake lock` |
| `orchestrator/internal/service/rules/worker/tool_executor.go` | Real tool scaffolding inside `nix develop` |
| `orchestrator/pkg/guids/` | Structured tool guide registry (commands, packages, structure) |
| `orchestrator/pkg/models/context_entry.go` | Context entry GORM model |
| `orchestrator/internal/service/context/` | Context management (Redis + Postgres, 3 scopes, auto-cleanup) |
| `projects/<taskID>/flake.nix` | Auto-generated per project |
| `internal/service/groupchat/` | Group Chat orchestration (star topology, concurrent rounds) |
| `internal/service/rules/worker/agent.go` | `WorkerAgent` — group chat participant with AI generation |

Run with NixOS:

```nix
{
  imports = [ octra.flake.nixosModules.octra-orchestrator ];
  services.octra-orchestrator = {
    enable = true;
    environment.DB_DNS = "postgres://...";
  };
}
```

## Quick start

```bash
docker compose up -d --build
```

Open http://localhost, describe your task, Octra builds it.  
Project is cached in Nix store — recover anytime via `RestoreProject(taskID)`.

---

*Octra: describe once, rebuild forever.*

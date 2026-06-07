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
| `orchestrator` | Go, gRPC | Boss → Manager → Worker pipeline, Nix snapshots, tool scaffolding |
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
| `projects/<taskID>/flake.nix` | Auto-generated per project |

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

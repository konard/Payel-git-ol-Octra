# Octra — AI agent factory powered by Nix

Octra is a multi-agent AI orchestrator that builds software projects using a Boss → Manager → Worker pipeline. Every project is snapshotted into the [Nix](https://nixos.org) store, enabling instant rollback and zero-storage recovery.

## Architecture (Boss → Manager → Worker)

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
**Workers** generate code inside Nix-isolated environments.

## Project lifecycle with Nix

```
setupProject() → generate flake.nix → AI generates code → nix-store --add → cleanup
                                                                 ↓
                                                          RestoreProject(taskID)
```

Each project:

1. Gets a `flake.nix` at creation time
2. Is built by AI agents inside the working directory
3. On completion: **snapshotted to `/nix/store/`** via `nix-store --add`
4. Working directory is removed (zero disk waste)
5. **`RestoreProject(taskID)`** recovers the project from the Nix store instantly

This means projects take zero space when idle but can be restored at any time.

## Services

| Service | Stack | Role |
|---------|-------|------|
| `orchestrator` | Go, gRPC | Boss → Manager → Worker pipeline, Nix snapshots |
| `agents` | Go, gRPC | AI provider proxy (Claude, Gemini, GPT, DeepSeek, ...) |
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

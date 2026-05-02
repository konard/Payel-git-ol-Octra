# CrewAI Project Context

## Current Status (May 2026)

### What Works
- Multi-agent system: Boss → Manager → Workers
- Project structure: projects/{taskID}/{projectName}/
- Git integration (.git in project folder)
- WebSocket streaming with redis
- File generation (Go, JS, Python code)
- **Frontend Code Editor** (CodeViewer) с подсветкой синтаксиса Kanagawa
- Worker count prediction based on grade_weight (1-20 → 1 worker, 21-40 → 1-2, etc.)

### Known Issues
1. **WebSocket disconnects**: Code 1006 - disconnect during manager review at ~87%
2. **LLM quality**: Worker generates truncated/bad code, Manager rejects
3. **TechStack ignored**: Boss decides Go but Worker creates JS - in progress fix
4. **Task doesn't complete**: Stops at "Manager backend reviewing workers..." (87%)
5. **File duplication**: AI adds project name to paths - fixed by using filepath.Base
6. **Username/user_id**: Was hardcoded "user" - now uses auth store

### Latest Changes (In Progress)
- worker/internal/prompts/task.go: Added techStack parameter to PlanFiles and GenerateFile functions
- worker/internal/service/worker/assign.go: Reading tech_stack from metadata, passing to generateCode
- worker/internal/service/worker/code.go: Added techStack parameter to generateCode function
- worker/internal/service/worker/code_multpass.go: Added techStack parameter, added language instruction in prompt
- boss/internal/service/boss/manager.go: Passing tech_stack in metadata to manager/worker
- apigateway: Added file streaming with 150ms delay for "busy activity" effect
- apigateway: Changed WebSocket ping from 25s to 20s
- Fixed JSON parsing fallback (extractJSONFromMarkdown finds JSON without markdown blocks)
- Added grade_weight system in boss metadata for worker count prediction
- manager: Added 30s timeout to reviewWorkerResult and managerThink AI calls (auto-approve on timeout)

### Architecture
```
User → Apigateway → Boss → Manager(s) → Workers → Git/Files
          ↓
       WebSocket (streaming)
          ↓
       Redis (history/reconnect)
```

### Project Storage
```
projects/
  {taskID}/
    {title}/         ← created by Boss
      .git/
      .crewai/
      main.go
      config.yaml
```

### Frontend Tech Stack
- React 18 + Vite
- Tailwind CSS 4
- highlight.js (syntax highlighting)
- @xyflow/react (React Flow)
- Radix UI components
- Zustand (state management)

### Services
| Service | Port | Language |
|---------|------|----------|
| Apigateway | 3111 | Go |
| Boss | 50051 | Go |
| Manager | 50052 | Go |
| Worker | 50053 | Go |
| Agents | 50053 | Go |
| Auth | 3112 | Go |
| Frontend | 80 | React/Vite |
| tgbot | - | Python |

### Environment Variables
- `DB_DNS` — PostgreSQL connection string
- `AGENTS_SERVICE_HOST` — Agents gRPC address
- `REDIS_URL` / `REDIS_ENABLED` — Redis for WebSocket and context caching
- `WORKER_MODE` — `multypass` (optimized) or `nplus1` (legacy)
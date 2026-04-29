# CrewAI Project Context

## Current Status (April 2026)

### What Works
- Multi-agent system: Boss → Manager → Workers
- Project structure: projects/{taskID}/{projectName}/
- Git integration (.git in project folder)
- WebSocket streaming with redis
- File generation (Go code)

### Known Issues
1. **WebSocket disconnects**: Code 1006 - Added ping every 25s to keep alive
2. **File duplication**: AI adds project name to paths - fixed by using filepath.Base
3. **Username/user_id**: Was hardcoded "user" - now uses auth store
4. **Translations missing**: Some keys not displaying properly

### Latest Changes
- apigateway: Added PingWriter for WebSocket keep-alive
- frontend: Use authUser.username and authUser.id for tasks
- manager: Removed title subfolder addition (Boss does it)
- worker: Uses only filename via filepath.Base
- worker: Unified file/command write path to a single `basePath` (`req.ProjectPath`/`repo_path`) to prevent writes outside task repo
- Boss: Creates project path with title subfolder

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
# Nix Project Restore via gRPC Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable switching between historical projects by restoring solution files from Nix store via a full gRPC chain: Frontend → User service → Orchestrator → Nix store → files back.

**Architecture:** The orchestrator already snapshots projects to Nix store and saves `nix_store_path` in the Task model. We add a new gRPC endpoint that restores a project from Nix, reads the code files, and returns them. The user service acts as a caching proxy (LRU cache of last 10 tasks, 500MB buffer). The frontend uses connect-web gRPC to call the user service directly.

**Tech Stack:** Protocol Buffers, gRPC-Go, connect-web (frontend), Nix store, LRU cache

---

## File Structure

| Service | File | Purpose |
|---------|------|---------|
| Orchestrator | `proto/gateway-boss.proto` | Add `RestoreProjectFiles` RPC + messages |
| Orchestrator | `internal/fetcher/grpc/bosspb/*.go` | Regenerated protobuf code |
| Orchestrator | `internal/fetcher/grpc/server.go` | Implement `RestoreProjectFiles` handler |
| Orchestrator | `internal/service/rules/boss/project.go` | Extract file reading logic |
| User service | `proto/user-files.proto` | New proto for user↔frontend gRPC |
| User service | `internal/grpc/client.go` | gRPC client to orchestrator |
| User service | `internal/grpc/server.go` | gRPC server for frontend |
| User service | `internal/cache/nix_cache.go` | LRU cache (10 tasks, 500MB) |
| User service | `pkg/models/chat.go` | Add `task_id`, `nix_store_path` fields |
| User service | `internal/core/services/chat.go` | Update chat service |
| Frontend | `src/grpc/client.ts` | connect-web gRPC client |
| Frontend | `src/grpc/user-files_pb.ts` | Generated protobuf |
| Frontend | `src/stores/taskStore.ts` | Add `loadProjectFiles` action |
| Frontend | `src/app/App.tsx` | Call gRPC on chat switch |

---

## Task 1: Proto Definition — Orchestrator

**Covers:** gRPC contract for project file restoration

**Files:**
- Modify: `orchestrator/proto/gateway-boss.proto`

- [ ] **Step 1: Add new messages and RPC to proto**

```protobuf
// Restore project files request
message RestoreProjectFilesRequest {
  string nix_store_path = 1;
  string task_id = 2;
}

// Single code file in the response
message CodeFileEntry {
  string path = 1;
  string content = 2;
  string language = 3;
  string encoding = 4;     // "base64" for binary files
  string worker_role = 5;
  string manager_role = 6;
}

// Restore project files response
message RestoreProjectFilesResponse {
  string task_id = 1;
  repeated CodeFileEntry files = 2;
  int32 total_files = 3;
}

// Add to BossService:
service BossService {
  // ... existing RPCs ...

  // Restore project files from Nix store snapshot
  rpc RestoreProjectFiles(RestoreProjectFilesRequest) returns (RestoreProjectFilesResponse);
}
```

- [ ] **Step 2: Regenerate protobuf code**

Run in `orchestrator/`:
```bash
protoc --go_out=. --go-grpc_out=. proto/gateway-boss.proto
```

Expected: New `RestoreProjectFiles` method in `bosspb.BossServiceClient` and `BossServiceServer` interfaces.

- [ ] **Step 3: Commit**

```bash
git add orchestrator/proto/gateway-boss.proto orchestrator/internal/fetcher/grpc/bosspb/
git commit -m "feat: add RestoreProjectFiles gRPC definition"
```

---

## Task 2: Orchestrator — Implement RestoreProjectFiles

**Covers:** Restore project from Nix store, read files, return via gRPC

**Files:**
- Modify: `orchestrator/internal/fetcher/grpc/server.go`
- Modify: `orchestrator/internal/service/rules/boss/project.go`

- [ ] **Step 1: Add file reading function to project.go**

```go
// ReadProjectFiles — читает файлы проекта из директории, фильтруя игнорируемые
func ReadProjectFiles(projectPath string) ([]streamedSolutionFile, error) {
    var files []streamedSolutionFile
    seen := map[string]bool{}

    err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil // skip unreadable files
        }
        if info.IsDir() {
            return nil
        }

        relPath, err := filepath.Rel(projectPath, path)
        if err != nil {
            return nil
        }
        relPath = filepath.ToSlash(relPath)

        if util.IsIgnoredPath(relPath) || seen[relPath] {
            return nil
        }
        seen[relPath] = true

        content, err := os.ReadFile(path)
        if err != nil {
            return nil
        }

        encoding := ""
        contentStr := string(content)
        if util.IsBinaryPath(relPath) {
            contentStr = base64.StdEncoding.EncodeToString(content)
            encoding = "base64"
        }

        files = append(files, streamedSolutionFile{
            Path:     relPath,
            Content:  contentStr,
            Language: util.LanguageForPath(relPath),
            Encoding: encoding,
            Status:   "ready",
        })
        return nil
    })

    return files, err
}
```

- [ ] **Step 2: Implement RestoreProjectFiles in server.go**

```go
// RestoreProjectFiles — восстанавливает проект из Nix store и возвращает файлы
func (s *Server) RestoreProjectFiles(ctx context.Context, req *bosspb.RestoreProjectFilesRequest) (*bosspb.RestoreProjectFilesResponse, error) {
    if req.NixStorePath == "" {
        return nil, fmt.Errorf("nix_store_path is required")
    }

    // Create temp dir for restore
    tmpDir, err := os.MkdirTemp("", "octra-restore-*")
    if err != nil {
        return nil, fmt.Errorf("failed to create temp dir: %w", err)
    }
    defer os.RemoveAll(tmpDir)

    // Restore from Nix store
    if err := s.boss.RestoreProjectFromNix(req.NixStorePath, tmpDir); err != nil {
        return nil, fmt.Errorf("failed to restore from nix store: %w", err)
    }

    // Read files
    codeFiles, err := boss.ReadProjectFiles(tmpDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read project files: %w", err)
    }

    // Convert to proto
    entries := make([]*bosspb.CodeFileEntry, 0, len(codeFiles))
    for _, f := range codeFiles {
        entries = append(entries, &bosspb.CodeFileEntry{
            Path:     f.Path,
            Content:  f.Content,
            Language: f.Language,
            Encoding: f.Encoding,
        })
    }

    return &bosspb.RestoreProjectFilesResponse{
        TaskId:     req.TaskId,
        Files:      entries,
        TotalFiles: int32(len(entries)),
    }, nil
}
```

- [ ] **Step 3: Add RestoreProjectFromNix to boss.Service**

```go
// RestoreProjectFromNix — публичный метод для восстановления из Nix store
func (s *Service) RestoreProjectFromNix(nixStorePath, destPath string) error {
    return s.restoreProjectFromStore(nixStorePath, destPath)
}
```

- [ ] **Step 4: Commit**

```bash
git add orchestrator/internal/fetcher/grpc/server.go orchestrator/internal/service/rules/boss/project.go
git commit -m "feat: implement RestoreProjectFiles gRPC handler"
```

---

## Task 3: User Service — Proto + gRPC Client to Orchestrator

**Covers:** User service gRPC client for orchestrator communication

**Files:**
- Create: `user/internal/grpc/client.go`
- Modify: `user/go.mod` (add grpc dependencies)

- [ ] **Step 1: Create gRPC client to orchestrator**

```go
package grpc

import (
    "context"
    "fmt"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// OrchestratorClient — gRPC клиент для связи с orchestrator
type OrchestratorClient struct {
    conn   *grpc.ClientConn
    client Bosspb_BossServiceClient
}

// NewOrchestratorClient подключается к orchestrator
func NewOrchestratorClient(address string) (*OrchestratorClient, error) {
    conn, err := grpc.Dial(address,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultCallOptions(
            grpc.MaxCallRecvMsgSize(500*1024*1024), // 500MB buffer
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
    }

    return &OrchestratorClient{
        conn:   conn,
        client: NewBossServiceClient(conn),
    }, nil
}

// RestoreProjectFiles запрашивает файлы проекта из Nix store
func (c *OrchestratorClient) RestoreProjectFiles(ctx context.Context, nixStorePath, taskID string) (*RestoreProjectFilesResponse, error) {
    return c.client.RestoreProjectFiles(ctx, &RestoreProjectFilesRequest{
        NixStorePath: nixStorePath,
        TaskId:       taskID,
    })
}

// Close закрывает соединение
func (c *OrchestratorClient) Close() error {
    return c.conn.Close()
}
```

- [ ] **Step 2: Copy proto and generate Go code**

Copy `orchestrator/proto/gateway-boss.proto` to `user/proto/` and generate:
```bash
cd user && protoc --go_out=. --go-grpc_out=. proto/gateway-boss.proto
```

- [ ] **Step 3: Commit**

```bash
git add user/internal/grpc/ user/proto/ user/go.mod user/go.sum
git commit -m "feat: add orchestrator gRPC client to user service"
```

---

## Task 4: User Service — LRU Cache

**Covers:** Cache last 10 tasks' files, 500MB buffer

**Files:**
- Create: `user/internal/cache/nix_cache.go`

- [ ] **Step 1: Implement LRU cache**

```go
package cache

import (
    "sync"
    "time"
)

type CacheEntry struct {
    TaskID    string
    Files     []CodeFileEntry
    Size      int64
    Accessed  time.Time
}

type NixCache struct {
    mu       sync.RWMutex
    entries  map[string]*CacheEntry
    order    []string // LRU order: most recent at end
    maxSize  int64    // 500MB
    maxItems int      // 10 tasks
}

func NewNixCache(maxSizeBytes int64, maxItems int) *NixCache {
    return &NixCache{
        entries:  make(map[string]*CacheEntry),
        order:    make([]string, 0, maxItems),
        maxSize:  maxSizeBytes,
        maxItems: maxItems,
    }
}

type CodeFileEntry struct {
    Path     string
    Content  string
    Language string
    Encoding string
}

func (c *NixCache) Get(taskID string) ([]CodeFileEntry, bool) {
    c.mu.Lock()
    defer c.mu.Unlock()

    entry, ok := c.entries[taskID]
    if !ok {
        return nil, false
    }

    entry.Accessed = time.Now()
    c.moveToEnd(taskID)
    return entry.Files, true
}

func (c *NixCache) Set(taskID string, files []CodeFileEntry) {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Calculate size
    var size int64
    for _, f := range files {
        size += int64(len(f.Content))
    }

    // Evict if needed
    for len(c.entries) >= c.maxItems || c.totalSize() + size > c.maxSize {
        c.evictOldest()
    }

    c.entries[taskID] = &CacheEntry{
        TaskID:   taskID,
        Files:    files,
        Size:     size,
        Accessed: time.Now(),
    }
    c.order = append(c.order, taskID)
}

func (c *NixCache) totalSize() int64 {
    var total int64
    for _, e := range c.entries {
        total += e.Size
    }
    return total
}

func (c *NixCache) evictOldest() {
    if len(c.order) == 0 {
        return
    }
    oldest := c.order[0]
    c.order = c.order[1:]
    delete(c.entries, oldest)
}

func (c *NixCache) moveToEnd(taskID string) {
    for i, id := range c.order {
        if id == taskID {
            c.order = append(c.order[:i], c.order[i+1:]...)
            c.order = append(c.order, taskID)
            return
        }
    }
}
```

- [ ] **Step 2: Commit**

```bash
git add user/internal/cache/
git commit -m "feat: add LRU cache for project files (10 tasks, 500MB)"
```

---

## Task 5: User Service — Chat Model Update

**Covers:** Store task_id and nix_store_path in Chat

**Files:**
- Modify: `user/pkg/models/chat.go` (already done)
- Modify: `user/internal/core/services/chat.go`

- [ ] **Step 1: Add UpdateChatNixPath service function**

```go
// UpdateChatNixPath сохраняет task_id и nix_store_path в чате
func UpdateChatNixPath(chatID, taskID, nixStorePath string) error {
    chatUUID, err := uuid.Parse(chatID)
    if err != nil {
        return errors.New("invalid chat ID")
    }

    result := database.Db.Model(&models.Chat{}).
        Where("id = ?", chatUUID).
        Updates(map[string]interface{}{
            "task_id":        taskID,
            "nix_store_path": nixStorePath,
            "updated_at":     time.Now(),
        })

    if result.Error != nil {
        return result.Error
    }

    invalidateChatCache(chatID)
    return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add user/internal/core/services/chat.go
git commit -m "feat: add UpdateChatNixPath service"
```

---

## Task 6: User Service — gRPC Server for Frontend

**Covers:** gRPC endpoint that frontend calls to get project files

**Files:**
- Create: `user/internal/grpc/server.go`
- Create: `user/proto/user-files.proto`

- [ ] **Step 1: Create proto for user↔frontend gRPC**

```protobuf
syntax = "proto3";

package userfiles;

option go_package = "user/internal/grpc/userfilespb";

message GetProjectFilesRequest {
  string chat_id = 1;
}

message CodeFileEntry {
  string path = 1;
  string content = 2;
  string language = 3;
  string encoding = 4;
}

message GetProjectFilesResponse {
  string chat_id = 1;
  string task_id = 2;
  repeated CodeFileEntry files = 3;
  int32 total_files = 4;
}

service UserFilesService {
  rpc GetProjectFiles(GetProjectFilesRequest) returns (GetProjectFilesResponse);
}
```

- [ ] **Step 2: Implement gRPC server**

```go
package grpc

import (
    "context"
    "fmt"

    "user/internal/cache"
    "user/internal/core/services"
    "user/internal/grpc/userfilespb"
)

type UserFilesServer struct {
    userfilespb.UnimplementedUserFilesServiceServer
    orchestrator *OrchestratorClient
    cache        *cache.NixCache
}

func NewUserFilesServer(orch *OrchestratorClient, c *cache.NixCache) *UserFilesServer {
    return &UserFilesServer{
        orchestrator: orch,
        cache:        c,
    }
}

func (s *UserFilesServer) GetProjectFiles(ctx context.Context, req *userfilespb.GetProjectFilesRequest) (*userfilespb.GetProjectFilesResponse, error) {
    // 1. Get chat from DB
    chat, err := services.GetChatByID(req.ChatId)
    if err != nil {
        return nil, fmt.Errorf("chat not found: %w", err)
    }

    if chat.NixStorePath == "" {
        return &userfilespb.GetProjectFilesResponse{
            ChatId: req.ChatId,
            Files:  []*userfilespb.CodeFileEntry{},
        }, nil
    }

    // 2. Check cache
    taskID := chat.TaskId
    if cached, ok := s.cache.Get(taskID); ok {
        entries := make([]*userfilespb.CodeFileEntry, 0, len(cached))
        for _, f := range cached {
            entries = append(entries, &userfilespb.CodeFileEntry{
                Path:     f.Path,
                Content:  f.Content,
                Language: f.Language,
                Encoding: f.Encoding,
            })
        }
        return &userfilespb.GetProjectFilesResponse{
            ChatId:     req.ChatId,
            TaskId:     taskID,
            Files:      entries,
            TotalFiles: int32(len(entries)),
        }, nil
    }

    // 3. Fetch from orchestrator
    resp, err := s.orchestrator.RestoreProjectFiles(ctx, chat.NixStorePath, taskID)
    if err != nil {
        return nil, fmt.Errorf("failed to restore files: %w", err)
    }

    // 4. Cache result
    cacheFiles := make([]cache.CodeFileEntry, 0, len(resp.Files))
    for _, f := range resp.Files {
        cacheFiles = append(cacheFiles, cache.CodeFileEntry{
            Path:     f.Path,
            Content:  f.Content,
            Language: f.Language,
            Encoding: f.Encoding,
        })
    }
    s.cache.Set(taskID, cacheFiles)

    // 5. Return
    entries := make([]*userfilespb.CodeFileEntry, 0, len(resp.Files))
    for _, f := range resp.Files {
        entries = append(entries, &userfilespb.CodeFileEntry{
            Path:     f.Path,
            Content:  f.Content,
            Language: f.Language,
            Encoding: f.Encoding,
        })
    }

    return &userfilespb.GetProjectFilesResponse{
        ChatId:     req.ChatId,
        TaskId:     taskID,
        Files:      entries,
        TotalFiles: resp.TotalFiles,
    }, nil
}
```

- [ ] **Step 3: Add GetChatByID to services**

```go
// GetChatByID — получить чат по ID (без проверки user_id)
func GetChatByID(chatID string) (*models.Chat, error) {
    id, err := uuid.Parse(chatID)
    if err != nil {
        return nil, errors.New("invalid chat ID")
    }

    var chat models.Chat
    err = database.Db.Where("id = ?", id).First(&chat).Error
    if err != nil {
        return nil, err
    }

    return &chat, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add user/internal/grpc/ user/proto/ user/internal/core/services/chat.go
git commit -m "feat: add gRPC server for frontend project files"
```

---

## Task 7: User Service — Wire Everything Together

**Covers:** Initialize gRPC server and client in main.go

**Files:**
- Modify: `user/cmd/app/main.go`

- [ ] **Step 1: Add gRPC server startup**

```go
// In main.go, after HTTP server setup:

// Initialize orchestrator gRPC client
orchAddr := os.Getenv("ORCHESTRATOR_GRPC_ADDR")
if orchAddr == "" {
    orchAddr = "localhost:50051"
}
orchClient, err := grpc.NewOrchestratorClient(orchAddr)
if err != nil {
    log.Printf("Warning: failed to connect to orchestrator: %v", err)
} else {
    defer orchClient.Close()

    // Initialize cache
    nixCache := cache.NewNixCache(500*1024*1024, 10) // 500MB, 10 items

    // Start gRPC server for frontend
    go func() {
        lis, err := net.Listen("tcp", ":50052")
        if err != nil {
            log.Fatalf("failed to listen: %v", err)
        }
        grpcServer := grpc.NewServer(
            grpc.MaxRecvMsgSize(500*1024*1024),
            grpc.MaxSendMsgSize(500*1024*1024),
        )
        userfilespb.RegisterUserFilesServiceServer(
            grpcServer,
            grpc.NewUserFilesServer(orchClient, nixCache),
        )
        log.Printf("user gRPC server listening on :50052")
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("failed to serve: %v", err)
        }
    }()
}
```

- [ ] **Step 2: Commit**

```bash
git add user/cmd/app/main.go
git commit -m "feat: wire gRPC server and orchestrator client in user service"
```

---

## Task 8: Frontend — connect-web gRPC Client

**Covers:** Frontend gRPC client for project files

**Files:**
- Create: `frontend/web/src/grpc/client.ts`
- Create: `frontend/web/proto/user-files.proto`
- Modify: `frontend/web/package.json` (add @connectrpc dependencies)

- [ ] **Step 1: Install dependencies**

```bash
cd frontend/web && npm install @connectrpc/connect @connectrpc/connect-web @bufbuild/protobuf
```

- [ ] **Step 2: Create proto file**

```protobuf
syntax = "proto3";

package userfiles;

message GetProjectFilesRequest {
  string chat_id = 1;
}

message CodeFileEntry {
  string path = 1;
  string content = 2;
  string language = 3;
  string encoding = 4;
}

message GetProjectFilesResponse {
  string chat_id = 1;
  string task_id = 2;
  repeated CodeFileEntry files = 3;
  int32 total_files = 4;
}

service UserFilesService {
  rpc GetProjectFiles(GetProjectFilesRequest) returns (GetProjectFilesResponse);
}
```

- [ ] **Step 3: Create gRPC client**

```typescript
// frontend/web/src/grpc/client.ts
import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { UserFilesService } from "./user-files_pb";

const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_GRPC_WEB_URL || "http://localhost:50052",
});

export const userFilesClient = createClient(UserFilesService, transport);
```

- [ ] **Step 4: Generate TypeScript code**

```bash
cd frontend/web
npx buf generate proto/user-files.proto --template buf.gen.yaml
```

- [ ] **Step 5: Commit**

```bash
git add frontend/web/src/grpc/ frontend/web/proto/ frontend/web/package.json
git commit -m "feat: add connect-web gRPC client for project files"
```

---

## Task 9: Frontend — Load Files on Chat Switch

**Covers:** Restore solution files when switching to historical chat

**Files:**
- Modify: `frontend/web/src/stores/taskStore.ts`
- Modify: `frontend/web/src/app/App.tsx`

- [ ] **Step 1: Add loadProjectFiles to taskStore**

```typescript
// In taskStore.ts, add new action:

interface TaskState {
  // ... existing ...
  loadProjectFiles: (chatId: string) => Promise<void>;
}

// In create():
loadProjectFiles: async (chatId: string) => {
  try {
    const { userFilesClient } = await import('../grpc/client');
    const response = await userFilesClient.getProjectFiles({ chatId });

    if (response.files.length > 0) {
      const codeFiles = response.files.map((f) => ({
        path: f.path,
        name: f.path.split('/').filter(Boolean).at(-1) || f.path,
        language: f.language || 'plaintext',
        encoding: f.encoding || undefined,
        content: f.content,
        status: 'ready' as const,
        updatedAt: Date.now(),
      }));

      set({
        codeFiles,
        latestCodeFilePath: codeFiles[codeFiles.length - 1]?.path ?? null,
      });
    }
  } catch (err) {
    console.error('Failed to load project files:', err);
  }
},
```

- [ ] **Step 2: Update handleSelectChat in App.tsx**

```typescript
const handleSelectChat = async (chatId: string) => {
  // ... existing code ...

  try {
    const chat = await getChat(chatId);
    setChatMessages(/* ... */);

    const { nodes, edges } = parseChatWorkflow(chat.workflow);
    useTaskStore.getState().setGraph(nodes, edges);

    // Load solution files from Nix store
    await useTaskStore.getState().loadProjectFiles(chatId);
  } catch (err) {
    console.error('Failed to load chat:', err);
  }
};
```

- [ ] **Step 3: Commit**

```bash
git add frontend/web/src/stores/taskStore.ts frontend/web/src/app/App.tsx
git commit -m "feat: load project files on chat switch via gRPC"
```

---

## Task 10: Save nix_store_path When Task Completes

**Covers:** Persist Nix store path to chat when task finishes

**Files:**
- Modify: `frontend/web/src/app/App.tsx`
- Modify: `frontend/web/src/services/chatHistoryService.ts`

- [ ] **Step 1: Add API to save nix path**

```typescript
// In chatHistoryService.ts:

export async function updateChatNixPath(
  chatId: string,
  taskId: string,
  nixStorePath: string
): Promise<void> {
  const response = await fetch(`${AUTH_API_URL}/chat/${chatId}/nix-path`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
    credentials: 'include',
    body: JSON.stringify({ task_id: taskId, nix_store_path: nixStorePath }),
  });

  if (!response.ok) {
    throw new Error('Failed to update nix path');
  }
}
```

- [ ] **Step 2: Save nix path from task completion data**

In `App.tsx`, when task completes and we receive the final update with `nix_store_path`:

```typescript
// In handleIncomingChatMessage or where task completion is handled:
if (data.nix_store_path && data.task_id && currentChatIdRef.current) {
  updateChatNixPath(
    currentChatIdRef.current,
    data.task_id,
    data.nix_store_path
  ).catch(console.error);
}
```

- [ ] **Step 3: Add backend route for nix-path**

In `user/internal/fetcher/http/router/chat/chat.go`:

```go
// PUT /chat/:id/nix-path
r.PUT("/chat/:id/nix-path", middleware.AuthMiddleware(), func(c *gin.Context) {
    chatID := c.Param("id")

    var req struct {
        TaskID       string `json:"task_id"`
        NixStorePath string `json:"nix_store_path"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"status": "error", "error": "Invalid request body"})
        return
    }

    err := services.UpdateChatNixPath(chatID, req.TaskID, req.NixStorePath)
    if err != nil {
        c.JSON(500, gin.H{"status": "error", "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "success"})
})
```

- [ ] **Step 4: Commit**

```bash
git add frontend/web/src/app/App.tsx frontend/web/src/services/chatHistoryService.ts user/internal/fetcher/http/router/chat/chat.go
git commit -m "feat: save nix_store_path to chat on task completion"
```

---

## Task 11: Send nix_store_path from Orchestrator

**Covers:** Orchestrator includes nix_store_path in final progress update

**Files:**
- Modify: `orchestrator/internal/service/rules/boss/task.go`

- [ ] **Step 1: Add nix_store_path to final emit**

In `ExecuteTask`, after `cleanupProject` (which saves nix_store_path to Task):

```go
// After cleanupProject saves nix_store_path:
if task.NixStorePath != "" {
    data["nix_store_path"] = task.NixStorePath
    data["task_id"] = taskID.String()
}
```

Note: Need to reload task after cleanupProject since it modifies the DB:

```go
s.cleanupProject(projectPath, taskID.String())

// Reload task to get nix_store_path saved by cleanupProject
database.Db.First(task, "id = ?", taskID)
```

- [ ] **Step 2: Commit**

```bash
git add orchestrator/internal/service/rules/boss/task.go
git commit -m "feat: include nix_store_path in final task update"
```

---

## Verification

After all tasks are complete:

1. **Proto compilation**: Both orchestrator and user service proto compile without errors
2. **Unit tests**: Cache eviction, file reading, proto serialization
3. **Integration test**: 
   - Create a task → complete it → verify nix_store_path saved to chat
   - Switch to that chat → verify files loaded from Nix store
   - Switch to another chat → verify cache eviction works
4. **Manual test**: Complete a task, switch to another chat, switch back — files should load instantly from cache

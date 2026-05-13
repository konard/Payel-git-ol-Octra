# Octra - Multi-Agent Code Generation Platform

<div align="center">
  <img src="octra-mascot.png" alt="Octra Mascot" width="200">
</div>

Octra is a modern web platform that automates software development using AI agents organized in a hierarchical IT company structure. Users describe projects in natural language, and the system generates complete, production-ready codebases published directly to GitHub repositories.

## Architecture

```
┌─────────────┐      HTTP/WS     ┌─────────────┐      gRPC       ┌─────────────┐
│   Frontend  │ ───────────────► │  Apigateway │ ──────────────► │    Boss     │
│   (React)   │     :80          │   (proxy)   │   :3111         │  (port 50051)│
└──────┬──────┘                  └──────┬──────┘                 └──────┬──────┘
       │                                 │                               │
       │                                 │                gRPC (PARALLEL)
       │                                 │               ┌───────┴───────┐
       │                                 ▼               ▼               ▼
       │                        ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
       │                        │ User Service│  │  Manager #1 │  │  Manager #2 │
       │                        │  (auth)     │  │  (port 50052)│  │  (port 50052)│
       │                        │   :3112     │  └──────┬──────┘  └──────┬──────┘
       │                        └─────────────┘         │                 │
       │                                                 │                 │
       │                                       gRPC (SEQUENTIAL)          │
       │                                                 ▼                 ▼
       │                                  ┌─────────────┐  ┌─────────────┐
       │                                  │   Worker(s) │  │   Worker(s) │
       │                                  │  (port 50053)│  │  (port 50053)│
       │                                  └──────┬──────┘  └──────┬──────┘
       │                                         │                 │
       │                                         ▼                 ▼
       │                                  ┌─────────────┐  ┌─────────────┐
       │                                  │   Agents    │  │   Agents    │
       │                                  │  (port 50053)│  │  (port 50053)│
       └─────────────────────────────────────────────────┼─────────────────┘
                                                         │
                                               LLM API (OpenRouter, Gemini,
                                               OpenAI, Claude, DeepSeek, Grok)
```

## What is Octra?

Octra revolutionizes software development by creating an AI-powered development team that operates like a real IT company. Instead of manually writing code, users simply describe their project requirements in natural language, and CrewAI handles the entire development lifecycle - from architectural planning to code implementation and deployment.

### How It Works:
1. **User** describes the task: "Create a REST API in Go with authentication"
2. **Boss Agent** analyzes requirements and designs the system architecture
3. **Manager Agents** hire and coordinate development teams
4. **Worker Agents** implement individual components and features
5. **System** validates code quality and publishes the complete project to GitHub

## Key Features

- **Intelligent Project Planning** - AI analyzes requirements and selects optimal technology stacks and architectures
- **Parallel Development Teams** - Multiple AI agents work simultaneously on different project components
- **Iterative Code Review** - Automated code validation and improvement cycles
- **GitHub Integration** - Direct publishing of generated codebases to GitHub repositories
- **Multi-Provider LLM Support** - Access to OpenRouter, Gemini, OpenAI, Claude, DeepSeek, and Grok models
- **Real-time Web Interface** - Live progress updates and interactive development canvas
- **Professional Code Quality** - Production-ready code with proper structure and documentation

## Quick Start

### 1. Clone and Setup
```bash
git clone <repository>
cd crewai
# Configure environment variables (API keys for LLM providers)
```

### 2. Launch with Docker Compose
```bash
docker compose up -d --build
```

### 3. Access the Application
- **Web Interface**: http://localhost
- **API Gateway**: http://localhost:3111
- **User Service**: http://localhost:3112

### 4. First Use
1. Open http://localhost
2. Register or sign in
3. Subscribe to Pro plan (test payments available for development)
4. Create a task using Canvas or Chat mode
5. Access your generated project on GitHub

## Pro Subscription

Octra operates on a subscription model to ensure service quality and accessibility.

### Pricing:
- **1 month** - $10
- **3 months** - $25 (17% savings)
- **6 months** - $50 (17% savings)
- **1 year** - $100 (17% savings)

### Pro Features:
- Access to all LLM providers (OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok)
- Unlimited task creation
- Priority processing
- Advanced code generation capabilities
- Custom providers and models
- Technical support

### Payments:
- Bank cards via YooKassa
- Automatic subscription renewal
- Secure payment processing
- Payment receipts and history

### Development:
- Test payments available for functionality verification
- Payment configuration via environment variables

## Project Structure

```
crewai/
├── frontend/web/        # React web application
│   ├── src/
│   │   ├── app/         # Main application
│   │   ├── components/  # UI components
│   │   ├── services/    # API services
│   │   ├── stores/      # Zustand state management
│   │   └── hooks/       # React hooks
│   ├── nginx.conf       # Nginx configuration
│   └── Dockerfile
├── user/                # User service (authentication, subscriptions)
│   ├── cmd/app/         # Main application
│   ├── internal/core/   # Business logic
│   ├── pkg/             # Helper packages
│   └── Dockerfile
├── apigateway/          # API Gateway
│   ├── cmd/app/
│   ├── internal/
│   └── pkg/
├── boss/                # Boss agent (coordinator)
├── manager/             # Manager agents
├── worker/              # Worker agents
├── agents/              # LLM providers
├── docker-compose.yml   # Docker configuration
├── go.work             # Go workspace
└── .env.example        # Environment variables example
```

## Development Workflow

```
1. User → Apigateway (WebSocket)
2. Apigateway → Boss (gRPC: CreateTaskStream)
3. Boss → Agents: "Analyze task, determine stack and managers"
4. Boss → Manager #1, #2, #3 (gRPC: AssignManager — PARALLEL)
   ├── Manager → Agents: "What workers to hire for my team?"
   ├── Manager → Worker(s) (gRPC: AssignWorkersAndWait)
   │   ├── Worker → Agents: "What files to create?"
   │   ├── Worker → Agents: "Write file 1"
   │   ├── Worker → Agents: "Write file 2"
   │   └── Worker → Code implementation
   ├── Manager → Agents: "Review worker output"
   ├── Manager → Worker (gRPC: ReviewWorker — if not approved)
   └── Manager → Code + results
5. Boss → Agents: "Validate final solution"
6. Boss → GitHub: Create repository and push code
7. Boss → Apigateway → User: GitHub repository URL
```

## Data Models

### Task
- ID, UserID, Username, Title, Description
- Tokens (JSON), Meta (JSON)
- Status: `pending → boss_planning → managers_assigned → processing → reviewing → done/error`
- ProjectJSON (GitHub repository URL)

### BossDecision
- TaskID, ManagersCount, ManagerRoles (JSON)
- TechnicalDescription, TechStack (JSON), ArchitectureNotes

### Manager
- TaskID, Role, Status
- WorkerRoles (JSON), WorkersCount

### Worker
- TaskID, ManagerID, Role, Status
- TaskMD, SolutionMD
- Files (JSON), Success, Approved, Feedback

## Development

### Requirements
- Docker & Docker Compose
- Go 1.25+
- Node.js 18+ (for frontend)
- PostgreSQL 15+
- Redis 7+

### Development Mode
```bash
# Clone
git clone <repository>
cd crewai

# Setup environment
# Edit environment variables

# Launch all services
docker compose up -d

# Or only required for development
docker compose up -d postgres redis user frontend
```

### Frontend Development
```bash
cd frontend/web
npm install
npm run dev  # Runs on http://localhost:5173
```

### Backend Development
```bash
# Local service startup (after DB launch)
cd user && go run cmd/app/main.go
cd apigateway && go run cmd/app/main.go
# ... other services
```

### Protobuf
```bash
# Generate Go code from proto files
# Use generation script or protoc manually
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  proto/*.proto
```

## Important Notes

1. **Pro Subscription Required** - Active Pro subscription needed for service usage
2. **LLM Provider API Keys** - Configure at least one LLM provider in environment variables
3. **PostgreSQL and Redis** - Used for data storage and caching
4. **Docker Compose** - Recommended for development and production deployment
5. **Test Payments** - Available for payment functionality testing without real money
6. **Performance** - Generation time depends on task complexity and selected model
7. **Security** - API keys stored securely and not logged

---

## Environment Variables

Create `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

### Required Variables:

```bash
# JWT tokens (generate random strings)
JWT_SECRET="your-jwt-secret-here"
JWT_REFRESH_SECRET="your-refresh-secret-here"

# YooKassa payments (test data)
YOOKASSA_SHOP_ID="1339826"
YOOKASSA_SECRET_KEY="test_StL4_VJfVFbOJ7_BbolU2VhoR1zjIQ7Qf2gcwN3Gngw"

# Provider API keys (at least one required)
OPENROUTER_API_KEY="sk-or-v1-..."
GEMINI_API_KEY="AIzaSy..."
# Other providers optional
```

### Optional Variables:

```bash
# Service ports (defaults)
AUTH_PORT=3112
API_GATEWAY_PORT=3111

# Rate limiting
RATE_LIMIT_TASK_CREATE=10/60
RATE_LIMIT_TASK_STATUS=60/60

# Redis (for caching)
REDIS_URL=redis://redis:6379/0

# GitHub token (for publishing generated code)
GITHUB_TOKEN=ghp_...
```
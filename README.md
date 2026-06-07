# Octra — AI-Powered Software Development Team

<div align="center">
  <img src="docs/icons/octra-mascot.png" alt="Octra Mascot" width="200">
</div>

**Describe your idea. Get production-ready code. Published to GitHub.**

Octra replaces an entire development team with AI agents. You describe what you want to build in natural language, and Octra handles everything from architecture planning to code generation to GitHub publishing — in minutes, not weeks.

## What Octra Solves

Building software is slow. You need to design architecture, set up projects, write hundreds of files, manage dependencies, fix bugs, and repeat. Octra automates the entire process:

- **From idea to code** — describe your task once, get a complete project
- **Multiple AI agents working together** — one plans architecture, others write code, review it, and fix issues
- **Any stack, any complexity** — REST APIs, microservices, frontends, CLIs, research reports, presentations
- **Ready to use** — generated project is pushed directly to your GitHub repository

## How It Works

```
Describe task → AI analyzes requirements → AI team builds it → Code on GitHub
```

## Key Features

- **Natural language input** — just describe what you need
- **Multi-agent AI pipeline** — architecture planning, parallel code generation, automated review, quality validation
- **Multiple LLM providers** — OpenRouter, Gemini, OpenAI, Claude, DeepSeek, Grok, Qwen, Z.AI
- **System skills** — agents read file-based expert skills (Research, Presentations, Frontend, Backend, DevOps, Proxy, VPN) and inject the matching guidance into their prompts (see `orchestrator/internal/skills/`)
- **GitHub integration** — results published directly to a new repository or pull request
- **Real-time progress** — watch your project being built step by step
- **Web interface** — interactive canvas and chat modes
- **Telegram bot** — create tasks from Telegram
- **Non-code output** — research reports, documentation, and slide decks with web-sourced references
- **Custom workflows** — save and reuse agent configurations

## Use Cases

- **Prototyping** — go from concept to working prototype in one session
- **Full-stack applications** — generate complete APIs with frontends, databases, authentication
- **Research & analysis** — produce well-sourced reports with web search
- **Presentations** — create slide decks with visual guidance and source attribution
- **Open-source contributions** — fix GitHub issues with automatically generated pull requests

## Get Started

```bash
docker compose up -d --build
```

Open http://localhost, describe your task, and Octra builds it.

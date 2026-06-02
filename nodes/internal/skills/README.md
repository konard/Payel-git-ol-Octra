# System Skills

This package gives Octra agents **file-based system skills** instead of only
hardcoded inline prompts. Each skill is a Markdown document describing a real
specialty (Research, Presentations, Frontend, Backend, DevOps, Proxy, VPN, …).
At runtime the agents *read* these skills, match the best one to the worker's
role / tech stack / task, and inject its expert guidance into the LLM prompt.

The library is embedded into the binary (`//go:embed library/*.md`), so skills
travel with the service and require no external network or filesystem access.

## How it is used

- **Boss planning** — `prompts.PlanArchitecture` lists the available skill areas
  (`skills.Catalog()`) so the boss picks manager/worker roles that map to a real
  specialty.
- **Workers** — code and document workers call `skills.Guidance(role, tech, task)`
  and append the returned expert block to their prompts (planning, file
  generation, research, document and presentation writing).

## Skill file format

Drop a new `library/<slug>.md` file with a small frontmatter block:

```markdown
---
name: Backend Development
slug: backend
area: Backend
keywords: backend, server, api, grpc, database, бэкенд
tech: go, golang, python, node
---
## Backend Development skill

...expert guidance injected into the prompt...
```

- `keywords` — matched against the worker role (weighted high) and task text.
- `tech` — matched against the worker's tech stack.
- The Markdown body after the frontmatter is what gets injected into prompts.

Adding a new specialty is just adding another `.md` file — no code changes.

## Sources / inspiration

The curated skill content is inspired by community skill and prompt
collections referenced in the original issue:
[FlowiseAI](https://github.com/FlowiseAI/Flowise),
[LangChain Hub](https://smith.langchain.com/hub),
[PromptBase](https://promptbase.com/) and
[awesome-chatgpt-prompts](https://github.com/f/awesome-chatgpt-prompts).

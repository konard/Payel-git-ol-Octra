---
name: DevOps Configuration & Secrets
slug: devops-config
area: Dev ops
keywords: config, secrets, environment, env, secret store
tech: docker, kubernetes, terraform, ansible, bash, yaml
weight: 5
---
- Configure via environment variables, never bake secrets into images.
- Use a secret store (Vault, K8s Secrets, AWS Secrets Manager).
- Never commit secrets to version control.

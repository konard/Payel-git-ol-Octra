# Contributing Guidelines (встраивается в user prompt)

Если в репозитории найден CONTRIBUTING.md, его содержимое вставляется целиком
в user prompt после секции с информацией о задаче, перед финальным `Proceed.`:

```
Issue to solve: {issueUrl}
...
{contributingGuidelines}  ← CONTRIBUTING.md вставляется сюда
...
Proceed.
```

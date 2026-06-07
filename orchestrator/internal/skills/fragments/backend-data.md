---
name: Backend Data Layer
slug: backend-data
area: Backend
keywords: database, sql, orm, migration, query, data
tech: go, golang, python, node, nodejs, java, rust, ruby, c#, php
weight: 5
---
- Use migrations for schema changes, parameterized queries (never string-concatenate SQL).
- Add indexes for hot paths, use transactions for multi-step writes.
- Repository pattern for data access abstraction.

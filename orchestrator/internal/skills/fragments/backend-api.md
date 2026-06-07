---
name: Backend API Design
slug: backend-api
area: Backend
keywords: api, rest, grpc, graphql, endpoint
tech: go, golang, python, node, nodejs, java, rust, ruby, c#, php
weight: 5
---
- API design: consistent, versioned endpoints; validate all input; return meaningful status codes and structured errors.
- Keep handlers thin — move business logic to service layer.
- RESTful: nouns for resources, HTTP verbs for actions. gRPC for internal service-to-service.

---
name: Proxy Headers & Forwarding
slug: proxy-headers
area: Proxy
keywords: headers, forwarding, x-forwarded-for, host, connect
tech: go, nginx, haproxy, envoy, python
weight: 5
---
- Preserve/append X-Forwarded-For, X-Forwarded-Proto, Host.
- Handle CONNECT tunneling for HTTPS.
- Strip hop-by-hop headers correctly.

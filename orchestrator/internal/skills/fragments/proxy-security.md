---
name: Proxy Security
slug: proxy-security
area: Proxy
keywords: security, ssrf, tls, whitelist, upstream
tech: go, nginx, haproxy, envoy, python
weight: 5
---
- Validate/whitelist upstream targets to prevent SSRF.
- Terminate or pass through TLS deliberately.
- Log method, target, status and latency without leaking sensitive headers.

---
name: VPN Cryptography
slug: vpn-crypto
area: Vpn
keywords: vpn, keys, crypto, private key, rotate, secrets
tech: go, wireguard, openvpn, bash, python
weight: 5
---
- Generate per-peer key pairs, never share private keys.
- Rotate credentials periodically.
- Store secrets outside the repo, use environment variables or secret store.

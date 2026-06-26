# User Balance Data Model

The monolith stores payment state directly on the `User` model and keeps an
append-only transaction ledger for audits.

## User Fields

| Field | Meaning |
|-------|---------|
| `balance` | Current credit balance. Can be negative in unlimited margin mode. |
| `margin_mode` | `unlimited` or `safe`. |
| `safe_margin_limit` | Minimum preserved balance for safe margin mode. |
| `auto_pay_interval` | `day`, `week`, `month`, `half_year`, or `year`. |
| `auto_pay_day` | Day selector for scheduled charges. |

## Agent Fields

| Field | Meaning |
|-------|---------|
| `priority` | Priority used by safe mode when deciding which agents remain active. |

## Transaction

| Field | Meaning |
|-------|---------|
| `id` | Transaction ID. |
| `user_id` | Owner account. |
| `type` | `credit` or `debit`. |
| `amount` | Credit amount moved. |
| `reason` | `registration`, `hosting`, `lefine_reward`, or `topup`. |
| `agent_id` | Optional related agent. |
| `balance_after` | Balance snapshot after the transaction. |
| `created_at` | Transaction creation timestamp. |

## UsageMetric

| Field | Meaning |
|-------|---------|
| `user_id` | Owner account. |
| `date` | Usage date. |
| `cpu_seconds` | CPU seconds consumed. |
| `memory_mb_hours` | Memory MB-hours consumed. |
| `disk_mb` | Disk MB footprint. |
| `load_percent` | User load relative to platform average. |

## Public Backend Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/billing/balance` | Read current balance and margin settings. |
| `PATCH` | `/billing/settings` | Update margin mode and auto-pay settings. |
| `GET` | `/billing/transactions` | List balance ledger entries. |
| `POST` | `/billing/topup` | Add credits from a top-up flow. |
| `POST` | `/billing/lefine-reward` | Add credits from a LeFine reward. |
| `POST` | `/billing/usage` | Record resource usage and debit hosting credits. |

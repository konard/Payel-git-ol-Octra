# User Payment Concept

Octra charges for active hosted environments, not for requests or tokens. A
user can create an AI CLI environment for an agent; while that environment is
active, resource usage is recorded and converted into credit debits.

## Credit Unit

- 1 credit = 10 cents.
- New users receive 100 registration credits.
- Credits are spent on hosting charges.
- Credits can be added by top-up flows or LeFine rewards.

## Hosting Charge

Usage is measured as CPU seconds, memory MB-hours, disk MB and a load percent.
The standard hosting charge is scaled by load:

```text
charge = standard_payment * load_percent / 100
```

When raw load and platform average load are available, the equivalent formula is:

```text
charge = standard_payment * user_load / average_load
```

Average load pays the standard amount. Lower-than-average load is cheaper;
higher-than-average load is more expensive.

## Margin Modes

### Unlimited

Unlimited margin is the default. Hosting charges are always debited and the
balance may become negative.

### Safe

Safe margin preserves `safe_margin_limit`. If a hosting charge would push the
balance below that limit, Octra suspends the user's current agent and does not
record the debit. Already-running chat is still governed by the agent's active
flag; new agents require a non-negative balance.

## Auto-Pay Cadence

Users store their preferred billing cadence on the account:

- day
- week
- month
- half_year
- year

`auto_pay_day` stores the selected day in the cadence. The current backend
persists these settings and applies charges when usage is posted; a scheduler
can use the same fields for future margin-call jobs.

# Hooks Architecture

## Current state (V3.1)

One hook: session metering. Implemented per-adapter via:
- `SessionHookEnabled() bool`  
- `Record(delta any)`

See `internal/agents/ADAPTER-CONTRACT.md` (TOPIC-07) for the interface.

## Planned hooks (not yet implemented)

When a concrete consumer needs them:
- **pre-task**: audit logging, budget cap enforcement
- **post-task**: progress aggregation, dashboard tick

## When to implement

Add a hook when you have TWO concrete consumers. One consumer = inline code.
Two consumers = extract to shared hook. Don't build before demand.

## Design constraints (for whoever implements)

- Hooks observe; they don't transform or cancel.
- Panics in hooks must be recovered; hooks never crash the main flow.
- Registration is global-in-process.
- Thread-safe (hooks may fire from streaming parsers).

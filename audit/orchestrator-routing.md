# Audit: Orchestrator Routing & Double Probe

## Probe Sequence (cold start to SDD delegation)
1. General Orchestrator (GO) Startup: 4+ probes (tool status).
2. Intent Resolution: Detects `/sdd-new`.
3. Delegation: GO launches SDD Orchestrator (SO).
4. SO Startup: 7+ probes (Session-Setup Triplet [3] + Tool Check [4]).

## General Orchestrator Probes
- Tool Availability Check: 4+ RPC calls
- Intent Resolution: 0 RPC calls (scanning)

## Overlap with SDD Orchestrator
- Session-Setup Triplet: 3 RPC calls
- Redundant/overlapping: ~4 RPC calls (Tool status probes duplicated)
- Fix: Router Gate to forward results of GO tool checks to SO in the `Task` prompt instead of re-probing.

## Verdict
MASTER-PLAN Phase 3 claim VALIDATED: Redundant probes found. 4+ redundant RPC calls detected.

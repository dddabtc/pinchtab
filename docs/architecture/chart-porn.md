# Chart Porn

## Overview

```
┌─────────────┐     HTTP      ┌──────────────┐      CDP       ┌───────────────┐
│   AI Agent  │ ────────────▶ │   Pinchtab   │ ─────────────▶ │    Chrome     │
│  (any LLM)  │ ◀──────────── │  (Go binary) │ ◀───────────── │ self-launched │
└─────────────┘    JSON/text  └──────────────┘   WebSocket    └───────────────┘
```

## Instance Manager

```
AI Agent / CLI / External Script
          ↓  (HTTP API calls to port 9867)
       ┌──────────────────────────┐
       │   HTTP Server + Router   │  ← cmd/pinchtab/main.go + api handlers
       └───────────┬──────────────┘
                   │
                   ↓
       ┌─────────────────────────┐
       │     Orchestrator        │  ← manages fleet, creates instances
       │  (instance allocation)  │
       └───────────┬─────────────┘
                   │  (for each instance)
                   ↓
   ┌───────────────┼───────────────┐
   │               │               │
   ↓               ↓               ↓
┌────────────┐   ┌─────────────┐   ┌─────────────┐
│ Instance 1 │   │ Instance 2  │   │ Instance N  │  ← isolated Chrome + profile
└────────────┘   └─────────────┘   └─────────────┘
         │               │               │
         └───────┬───────┼───────┬───────┘
                 │       │       │
                 ↓       ↓       ↓
         ┌──────────────────────────────┐
         │         Bridge Layer         │  ← per-instance CDP client + stealth
         │  (CDP WebSocket + injection) │
         └──────────────────────────────┘
                 │       │       │
                 ↓       ↓       ↓
           Real Chrome Processes
           (headless / headed)
```

### Strategy Allocator

```
Incoming HTTP Request (e.g. /navigate?profile=agent1)
          ↓
    Orchestrator Handler
          ↓ (lookup)
    Instance exists for profile? ── Yes ──→ Use existing (bridge already connected)
          │ No
          ↓
    Allocator: CanAllocate() ?
      ├── Yes ── Reserve port + profile dir
      │           ↓
      │     Launch Chrome subprocess (with hardened flags)
      │           ↓
      │     Bridge: Connect CDP → Register instance → Success 200
      └── No (over limit / no resources)
                 ↓
           Return 429 / queue / wait-for-reclamation
```

## TabManager

```
External (AI / CLI / HTTP)
          ↓
    HTTP API Layer (handlers, router)
          ↓
    Orchestrator ── allocates → Instance (Chrome process + profile)
                            │
                            ↓   (per instance)
                  Instance Manager / Launcher
                            │
                            ↓
                ┌─────────────────────────────┐
                │       TabManager            │   ← the layer we're digging into
                │   (tabs map[string]*Tab)    │
                └───────────────┬─────────────┘
                                │   (for each tab)
                   ┌────────────┼────────────┐
                   │            │            │
                   ↓            ↓            ↓
             TabEntry 1    TabEntry 2    TabEntry N
               │               │               │
               └───────┬───────┼───────┬───────┘
                       │       │       │
                       ↓       ↓       ↓
                 Per-Tab CDP Bridge / Session
                       │       │       │
                       └───────┼───────┘
                               ↓
                       Chrome Process (CDP WebSocket)
```


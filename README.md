# TELCO-AI

Multi-agent network intelligence platform for telecom infrastructure — topology awareness, anomaly detection, fault correlation, and QKD (Quantum Key Distribution) link health, orchestrated over a **gRPC + NATS JetStream** backbone with a custom **GN (GenNets) engine** for graph learning and synthetic data generation.

## Repository layout

```
TELCO-AI/
├── ARCHITECTURE.md                          # System architecture (v0.1 draft)
├── TELCO_AI_Agent_Paper.docx                # Design paper
├── Telecom-Digital-Twin-Architecture_1.docx # Digital-twin architecture write-up
└── telco-ai-agent-scaffold/
    └── telco-ai-agent/                      # Project scaffold (proto, services, gn-engine, frontend, deploy, docs)
```

## Architecture at a glance

| Component | Lang | Role |
|---|---|---|
| Orchestrator | Go | gRPC gateway, session/state registry, saga-style coordination, token rotation |
| Topology SubAgent | Go | live network graph from telemetry, publishes deltas/snapshots |
| Anomaly SubAgent | Python | runs GN engine inference, scores nodes/edges for anomaly likelihood |
| Fault Correlation SubAgent | Go | groups graph/time-adjacent anomalies into ranked fault hypotheses |
| QKD Health Agent | Go | tracks QBER, key rate, channel drift on quantum links |
| GN Engine | lib | GNN (message-passing) + synthetic generator ("GenNets") |
| Frontend | Next.js | D3 topology map, Recharts metrics, Kanban fault board, QKD panel |

- No shared memory between agents — cross-agent traffic flows over **gRPC** (sync) or **NATS JetStream** (async, durable).
- Single `.proto` source of truth, generated for both Go and Python.
- **Postgres/TimescaleDB** stores metrics, anomaly events, faults, QKD health, topology snapshots.
- The simulator is the GN engine's live-mode generator wired to the real message bus — the same dashboard path runs for synthetic or live data.

See [`ARCHITECTURE.md`](./ARCHITECTURE.md) for the full design, and the write-ups under `telco-ai-agent-scaffold/telco-ai-agent/docs/`.

## Status

v0.1 draft — contracts and scaffold in progress. See `ARCHITECTURE.md` §9 for the suggested build order.
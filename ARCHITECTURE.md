# TELCO AI Agent — System Architecture

**Version:** 0.1 (draft)
**Scope:** Multi-agent network intelligence platform for telecom infrastructure — topology awareness, anomaly detection, fault correlation, and QKD (Quantum Key Distribution) link health, orchestrated over a gRPC + NATS JetStream backbone, with a custom GN (GenNets) engine for graph learning and synthetic data generation.

---

## 1. System Overview

```
                                   ┌─────────────────────────────┐
                                   │        Next.js Frontend      │
                                   │  React + Recharts + D3       │
                                   │  (Topology map, timelines,    │
                                   │   fault board, QKD panel)     │
                                   └───────────────┬──────────────┘
                                                   │ gRPC-Web / WS
                                   ┌───────────────▼──────────────┐
                                   │         Orchestrator          │
                                   │  Go · gRPC server · NATS      │
                                   │  client · session/state mgmt  │
                                   └───┬───────┬───────┬──────┬────┘
                     NATS JetStream    │       │       │      │      gRPC (control/query)
                 (telemetry, events,   │       │       │      │
                  tokens, streams)     │       │       │      │
        ┌──────────────────────────────┘       │       │      └───────────────────────┐
        │                              ┌────────┘       └────────┐                     │
┌───────▼────────┐            ┌────────▼───────┐        ┌────────▼────────┐   ┌────────▼────────┐
│ Topology        │            │ Anomaly         │        │ Fault Corr       │   │ QKD Health       │
│ SubAgent (Go)   │◄──gRPC────►│ SubAgent (Py)   │◄─gRPC─►│ SubAgent (Go)    │   │ Agent (Go)       │
│                 │            │                 │        │                  │   │                  │
│ - graph builder │            │ - GN engine     │        │ - correlates     │   │ - QBER, key rate │
│ - link/node     │            │   inference      │        │   anomalies →    │   │ - sifting ratio  │
│   state         │            │ - scoring        │        │   root cause     │   │ - channel drift  │
└───────┬─────────┘            └────────┬─────────┘        └────────┬─────────┘   └────────┬─────────┘
        │                                │                           │                       │
        └────────────────────────────────┴──────────┬────────────────┴───────────────────────┘
                                                      │
                                          ┌───────────▼────────────┐
                                          │   Postgres / Timescale  │
                                          │   (hypertables: metrics,│
                                          │    events, topology     │
                                          │    snapshots, faults)   │
                                          └─────────────────────────┘
```

Each subagent is an independently deployable Go (or Python) microservice. There is no shared memory — all cross-agent communication happens over:

1. **gRPC** — synchronous, typed request/response (queries, control commands, model inference calls).
2. **NATS JetStream** — asynchronous, durable streams (telemetry ingestion, event fan-out, token/credential propagation, replay for late subscribers).

---

## 2. Components

### 2.1 Orchestrator (Go)

Responsibilities:
- Owns the gRPC gateway exposed to the frontend (via grpc-gateway or Connect, translated to gRPC-Web).
- Maintains a **session/state registry**: which subagents are alive, their last heartbeat, current topology version, active fault tickets.
- Publishes control messages and orchestration commands (e.g. "recompute topology", "re-run correlation window") onto NATS subjects.
- Performs **saga-style coordination**: e.g. when Anomaly SubAgent flags a score above threshold, Orchestrator triggers Fault Corr SubAgent with a bounded time window and correlation ID.
- Issues and rotates internal **service tokens** (short-lived JWT or NATS NKey-based) distributed via a dedicated NATS subject (`telco.auth.tokens`), so subagents never need static shared secrets.

Key packages: `orchestrator/internal/registry`, `orchestrator/internal/saga`, `orchestrator/internal/gwserver`.

### 2.2 Topology SubAgent (Go)

- Ingests raw discovery/telemetry (SNMP/streaming telemetry/NetConf-derived events, or synthetic feed in simulation mode) from NATS subject `telco.telemetry.raw.*`.
- Builds and maintains a live graph model: nodes = network elements (routers, ONTs, OLTs, QKD nodes, cell sites), edges = physical/logical links with attributes (bandwidth, latency, loss, link type).
- Publishes topology deltas to `telco.topology.delta` and full snapshots periodically to `telco.topology.snapshot`, and writes versioned snapshots to Postgres.
- Exposes gRPC `TopologyService`: `GetGraph`, `GetNodeNeighborhood`, `SubscribeDeltas` (server-streaming), `GetGraphAtTime` (for time travel/replay).

### 2.3 Anomaly SubAgent (Python)

- Subscribes to `telco.telemetry.metrics.*` (JetStream, consumer with durable name per replica for horizontal scaling).
- Runs the **GN engine** (see §4) in inference mode: a graph neural network scores nodes/edges for anomaly likelihood using the current topology graph as structural context plus a sliding window of time-series features.
- Publishes scored anomalies to `telco.anomaly.events` with a correlation ID, severity, affected node/edge set, and a feature-attribution summary (which signals drove the score — for explainability).
- Exposes gRPC `AnomalyService`: `ScoreWindow`, `GetActiveAnomalies`, `StreamAnomalies`.
- Because it's the one Python service in an otherwise-Go fleet, it's isolated behind a strict gRPC contract (defined once in `.proto`, generated for both languages) — no cross-language shared state.

### 2.4 Fault Correlation SubAgent (Go)

- Subscribes to `telco.anomaly.events` and `telco.topology.delta`.
- Applies temporal + topological correlation: groups anomalies that are graph-adjacent and time-adjacent into a single **fault hypothesis**, ranks candidate root causes (e.g. shared upstream link, common power domain, correlated QKD channel degradation).
- Maintains a rolling correlation window (configurable, default 5 min) using an in-memory interval tree keyed by topology distance.
- Publishes fault tickets to `telco.fault.tickets` and persists to Postgres `faults` hypertable.
- Exposes gRPC `FaultCorrService`: `GetOpenFaults`, `GetFaultDetail`, `SubscribeFaults`.

### 2.5 QKD Health Agent (Go)

- Specialized health monitor for quantum key distribution links: tracks QBER (quantum bit error rate), sifted/final key rate, channel drift, and detector dead-time statistics.
- Treats QKD links as a distinct edge-type in the topology graph so degradation can participate in general fault correlation (e.g. a QKD channel fault correlating with a classical-channel fiber fault on the same physical route).
- Publishes to `telco.qkd.health` and raises anomaly-compatible events onto `telco.anomaly.events` so the Fault Corr SubAgent doesn't need QKD-specific logic.
- Exposes gRPC `QKDHealthService`: `GetLinkHealth`, `StreamLinkHealth`, `GetKeyRateHistory`.

### 2.6 GN (GenNets) Engine — shared library

A custom library (not a microservice — vendored/imported by Anomaly SubAgent and used offline for training) with two halves:

1. **GNN inference/training core** — graph neural network (message-passing, e.g. GraphSAGE/GAT-style layers) operating over the live topology graph, producing per-node and per-edge anomaly scores plus embeddings usable by Fault Corr for similarity-based grouping.
2. **Synthetic generator ("GenNets")** — a parametric generator that produces realistic synthetic topology + telemetry + fault-injection scenarios, used for (a) training the GNN before real data is abundant, (b) the simulator/demo environment, and (c) chaos-style testing of the whole pipeline.

See §4 for detail — this is the most novel piece of the system and is written up as its own design note.

### 2.7 Frontend (Next.js + React)

- **Topology view**: force-directed / hierarchical graph (D3) with live delta updates over a WebSocket bridge from the Orchestrator's gRPC-Web streaming endpoint.
- **Anomaly & metrics view**: Recharts time series per node/edge, overlaid anomaly score bands.
- **Fault board**: Kanban-style list of open fault tickets with root-cause graph highlight.
- **QKD panel**: QBER/key-rate gauges and history per quantum link.
- All views subscribe to the same delta stream abstraction so the UI never diverges from backend state (no separate polling paths).

---

## 3. Communication Contracts

### 3.1 NATS JetStream subjects (initial set)

| Subject | Producer | Consumers | Durability |
|---|---|---|---|
| `telco.telemetry.raw.>` | collectors / simulator | Topology SubAgent | WorkQueue, 24h retention |
| `telco.telemetry.metrics.>` | Topology SubAgent (enriched) | Anomaly SubAgent | Interest, 24h |
| `telco.topology.delta` | Topology SubAgent | Fault Corr, Frontend bridge | Interest, 7d |
| `telco.topology.snapshot` | Topology SubAgent | all, cold start | Interest, 30d |
| `telco.anomaly.events` | Anomaly SubAgent, QKD Health | Fault Corr, Frontend bridge | Interest, 7d |
| `telco.qkd.health` | QKD Health Agent | Frontend bridge, Fault Corr | Interest, 7d |
| `telco.fault.tickets` | Fault Corr SubAgent | Frontend bridge, Orchestrator | Interest, 30d |
| `telco.auth.tokens` | Orchestrator | all subagents | Interest, short TTL |

Stream names mirror subject prefixes (`TELEMETRY`, `TOPOLOGY`, `ANOMALY`, `QKD`, `FAULT`, `AUTH`), each its own JetStream stream so retention/replay policy can differ per domain.

### 3.2 gRPC services

Each subagent owns one `.proto` file under `proto/`, versioned independently (`telco.topology.v1`, `telco.anomaly.v1`, etc.), compiled for Go and Python from a single source of truth to avoid drift. The Orchestrator's gateway proto composes these into the external-facing API.

---

## 4. GN Engine — Design Note

### 4.1 GNN component

- **Input**: current topology graph (nodes/edges with static + dynamic attributes) + a rolling window of per-node/edge time-series (utilization, error rate, latency, QBER where applicable).
- **Model**: message-passing GNN (2–3 layers), attention-weighted edge aggregation so the model can learn "which neighbors matter" per node type (a QKD node's meaningful neighbors differ from a core router's).
- **Output**: per-node/edge anomaly score in [0,1] + embedding vector (used downstream by Fault Corr for clustering similar-looking anomalies even across different physical causes).
- **Training loop**: offline, versioned model artifacts, served via a lightweight inference server inside the Anomaly SubAgent process (no separate model-serving hop for v0.1 — revisit if latency/scale demands it).

### 4.2 Synthetic generator ("GenNets")

- Parametrized by: network scale (node/edge counts), topology archetype (ring, mesh, hierarchical access/aggregation/core, hybrid with QKD overlay), traffic profile, and a **fault-injection schedule** (link degradation, node flap, QKD channel drift, correlated multi-fault scenarios).
- Two output modes:
  1. **Batch mode** — writes a full synthetic dataset (topology snapshot + time-series + labeled fault windows) for offline GNN training/evaluation.
  2. **Live mode** — streams synthetic telemetry onto the real NATS subjects in real time, so the entire pipeline (Topology → Anomaly → Fault Corr → QKD → Frontend) can run against generated data. This live mode *is* the simulator described in §5.
- Labeled ground truth (which nodes/edges are actually anomalous, and why) is preserved alongside the synthetic stream so model quality and correlation accuracy can be scored automatically.

---

## 5. Simulator

The simulator is the GN engine's live-mode generator wired directly into the real message bus, plus a frontend that renders it exactly as it would render production data. This is deliberate: the same dashboard code path is exercised whether the data is synthetic or live, which also makes the simulator a standing integration test for the whole system.

Two granularities are useful during development:

- **In-process/browser mock** (fastest iteration): a self-contained frontend-only simulation for UI/UX work, generating synthetic topology + anomalies + faults client-side with no backend dependency.
- **Full-stack simulation** (integration-grade): the Go `gn-sim` CLI/service publishing onto real NATS subjects, exercised by the actual Go/Python subagents end-to-end, with Postgres/Timescale persistence — this is what CI and demos should run against.

---

## 6. Data Model (Postgres / TimescaleDB)

Core hypertables (chunked on `time`):

- `metrics(time, node_id, edge_id, metric_name, value, tags jsonb)`
- `anomaly_events(time, correlation_id, node_ids[], edge_ids[], severity, score, attribution jsonb)`
- `faults(id, opened_at, closed_at, root_cause_node_id, hypothesis jsonb, status)`
- `qkd_health(time, link_id, qber, key_rate_bps, sifting_ratio, drift)`
- `topology_snapshots(version, time, graph jsonb)` (or normalized `nodes`/`edges` tables with a `topology_version` column, depending on query patterns — start jsonb for velocity, normalize once query patterns stabilize)

Continuous aggregates (Timescale) for rollups feeding the frontend's longer time ranges (hourly/daily anomaly rates, QBER trends).

---

## 7. Deployment Topology (initial)

- Single `docker-compose` for local dev: NATS (JetStream enabled), Postgres+Timescale extension, all Go services, the Python anomaly service, and the Next.js frontend.
- Each subagent gets its own Dockerfile; Orchestrator is the only one with an external-facing port besides the frontend.
- Kubernetes manifests deferred until the local/dev loop is solid — premature to design pod topology before the contracts (proto + NATS subjects) are proven out.

---

## 8. Open Design Questions

1. **Model serving boundary** — keep the GNN in-process with the Anomaly SubAgent, or split into a dedicated inference service once multiple consumers need it (e.g. Fault Corr wanting embeddings directly)?
2. **Exactly-once vs at-least-once** — JetStream gives at-least-once by default; do any consumers (e.g. fault ticket creation) need idempotency keys to avoid duplicate tickets on redelivery? (Leaning yes — correlation ID + upsert.)
3. **QKD data realism** — how detailed should the synthetic QKD channel model be (BB84-style QBER dynamics vs. a simpler stochastic degradation model)? Affects both the generator and how meaningful QKD Health's output is.
4. **Multi-tenancy** — is this single-network or multi-operator from day one? Affects subject naming (`telco.<tenant>.topology.delta`) and Postgres partitioning.

---

## 9. Suggested Build Order

1. Proto contracts + NATS subject/stream definitions (frozen early, versioned after).
2. Topology SubAgent + Postgres schema + browser-mock simulator for the frontend team to start against immediately.
3. Orchestrator skeleton (registry + gateway, no saga logic yet).
4. Anomaly SubAgent with a placeholder scorer (random/threshold) — wire the pipeline end-to-end before the GNN is real.
5. GN engine v0: synthetic generator first (unblocks everyone), GNN model second.
6. Fault Corr SubAgent.
7. QKD Health Agent.
8. Full-stack `gn-sim` integration simulation replacing the browser mock as the default dev/demo mode.

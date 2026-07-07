# TELCO AI Agent

Multi-agent network intelligence platform: topology awareness, anomaly
detection, fault correlation, and QKD link health, orchestrated over a
gRPC + NATS JetStream backbone, with a custom GN (GenNets) engine for
graph learning and synthetic data/simulation.

See **docs/ARCHITECTURE.md** for the full design.

## Layout

```
proto/                  gRPC contracts (source of truth, Go + Python codegen)
services/
  orchestrator/          Go — gateway, subagent registry, saga coordination
  topology-agent/        Go — live network graph
  anomaly-agent/         Python — GN engine GNN inference
  faultcorr-agent/       Go — temporal + topological correlation
  qkd-health-agent/      Go — QKD link health monitoring
gn-engine/               Shared GN (GenNets) library
  gnn/                    Go reference scorer + shared inference contract
  gensynth/               Synthetic topology/telemetry/fault generator
gn-sim/                  CLI: drives gensynth in live mode onto real NATS subjects
frontend/                Next.js + React + Recharts/D3
deploy/
  docker/                docker-compose.yml (local dev stack)
  migrations/             TimescaleDB schema
docs/
  ARCHITECTURE.md         Full system design
```

## Quickstart (local dev)

```bash
# 1. Bring up NATS JetStream + TimescaleDB + all services
cd deploy/docker && docker compose up --build

# 2. Or, work on one Go service at a time (workspace-aware):
go work sync
cd services/topology-agent && go run .

# 3. Python anomaly agent
cd services/anomaly-agent
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
python main.py

# 4. Frontend
cd frontend && npm install && npm run dev

# 5. Full-stack simulator (synthetic topology + telemetry over real NATS)
cd gn-sim && go run . --nodes=80 --archetype=hybrid_qkd
```

## Build order

See docs/ARCHITECTURE.md §9. Short version: proto contracts first,
then Topology SubAgent + browser-mock simulator to unblock frontend
work immediately, then Orchestrator skeleton, then Anomaly SubAgent
with a placeholder scorer to prove the pipeline end-to-end before the
GNN is real, then GN engine (gensynth first, GNN model second), then
Fault Corr, then QKD Health, then swap the browser mock for the
full-stack `gn-sim`.

## Status

Scaffold stage: proto contracts defined, Go module skeletons in place,
`gensynth` has a working v0 topology/fault generator, all services
have TODO-marked wiring for gRPC + NATS. Nothing here talks to NATS
yet — that's the next milestone.

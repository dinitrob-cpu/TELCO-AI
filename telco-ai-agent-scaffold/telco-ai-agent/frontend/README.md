# Frontend (Next.js + React)

Views, matched to §2.7 of docs/ARCHITECTURE.md:

- `app/topology/` — force-directed / hierarchical graph (D3), live delta updates
- `app/anomalies/` — Recharts time series + anomaly score overlays
- `app/faults/` — fault ticket board with root-cause graph highlight
- `app/qkd/` — QBER / key-rate gauges and history per quantum link

All views should subscribe to one shared delta-stream client
(`lib/deltaClient.ts`, TODO) rather than polling independently, so the
UI never diverges from backend/simulator state.

For fast UI iteration without the Go/NATS stack running, see the
standalone browser-mock simulator artifact (no backend dependency).

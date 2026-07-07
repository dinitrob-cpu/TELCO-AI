# gn-engine

The GN (GenNets) engine: shared library used by the Anomaly SubAgent
and the simulator.

Two halves, two packages:

- `gnn/` — graph neural network inference core (message-passing GNN
  scoring nodes/edges for anomaly likelihood against the live topology
  graph + a sliding telemetry window).
- `gensynth/` — synthetic topology + telemetry + fault-injection
  generator ("GenNets"), used both for offline GNN training and for
  the live-mode simulator (`gn-sim`).

This is a Go library primarily for the `gensynth` generator (consumed
by `gn-sim` and the Go subagents). The GNN inference core used in
production by the Python Anomaly SubAgent is implemented in Python
(see `services/anomaly-agent/gn_engine/`) — the Go `gnn/` package here
defines the shared *interfaces and data contracts* (graph + feature
window in, scores + embeddings out) so both language implementations
stay contract-compatible, and hosts a pure-Go reference/fallback
scorer for use in `gn-sim` without a Python dependency.

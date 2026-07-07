"""Message-passing GNN for per-node/edge anomaly scoring.

v0: a minimal GraphSAGE-style model. Input is the current topology
graph (from Topology SubAgent, via gRPC) plus a rolling per-node/edge
telemetry window; output is an anomaly score in [0, 1] per node/edge
plus an embedding used downstream by Fault Corr SubAgent for
similarity-based grouping of anomalies.
"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass
class NodeFeatures:
    node_id: str
    window: list[float]


@dataclass
class EdgeFeatures:
    edge_id: str
    window: list[float]


@dataclass
class Score:
    id: str
    score: float
    embedding: list[float]


class GNNScorer:
    """Thin wrapper around the trained torch_geometric model.

    TODO: load a trained checkpoint; this stub returns zero scores so
    the service is runnable (and testable end-to-end over gRPC) before
    a trained model exists.
    """

    def __init__(self, checkpoint_path: str | None = None):
        self.checkpoint_path = checkpoint_path
        self.model = None  # TODO: torch_geometric model, loaded from checkpoint

    def score(
        self, nodes: list[NodeFeatures], edges: list[EdgeFeatures]
    ) -> list[Score]:
        out: list[Score] = []
        for n in nodes:
            out.append(Score(id=n.node_id, score=0.0, embedding=[]))
        for e in edges:
            out.append(Score(id=e.edge_id, score=0.0, embedding=[]))
        return out

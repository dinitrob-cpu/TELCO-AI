"""Entrypoint for the Anomaly SubAgent (Python).

Subscribes to telco.telemetry.metrics.> on NATS JetStream, runs the
GN engine GNN scorer using the current topology graph as structural
context, and publishes scored anomalies to telco.anomaly.events. Also
serves AnomalyService over gRPC (see proto/anomaly/anomaly.proto).
"""
import logging

from gn_engine.model import GNNScorer

logging.basicConfig(level=logging.INFO)
log = logging.getLogger("anomaly-agent")


def main() -> None:
    scorer = GNNScorer()
    log.info("anomaly-agent: starting (scorer=%s)", scorer.__class__.__name__)
    # TODO: connect to NATS JetStream, subscribe telco.telemetry.metrics.>,
    # maintain topology graph client, serve AnomalyService gRPC.


if __name__ == "__main__":
    main()

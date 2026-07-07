"""GN (GenNets) engine — GNN inference core, Python implementation.

This is the production scorer consumed by the Anomaly SubAgent over
the gRPC contract defined in proto/anomaly/anomaly.proto. It mirrors
the data contract in gn-engine/gnn (Go) so the two implementations
stay interchangeable — the Go reference scorer is used by the
simulator; this one is used in the real pipeline.
"""

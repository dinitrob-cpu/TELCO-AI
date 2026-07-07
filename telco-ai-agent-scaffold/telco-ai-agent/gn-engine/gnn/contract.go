// Package gnn defines the shared inference contract for the GN engine's
// graph neural network component. The production implementation lives
// in Python (services/anomaly-agent/gn_engine/); this package also
// ships a pure-Go reference scorer used by the simulator so it can run
// without a Python dependency.
package gnn

type NodeFeatures struct {
	NodeID   string
	Window   []float64 // rolling metric window, model-defined ordering
}

type EdgeFeatures struct {
	EdgeID string
	Window []float64
}

type Score struct {
	ID        string // node or edge id
	Score     float64
	Embedding []float64
}

// Scorer is implemented by both the Go reference model and (via gRPC,
// see proto/anomaly) the production Python GNN.
type Scorer interface {
	Score(nodes []NodeFeatures, edges []EdgeFeatures) ([]Score, error)
}

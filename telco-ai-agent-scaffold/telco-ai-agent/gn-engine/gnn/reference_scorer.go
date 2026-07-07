package gnn

// ReferenceScorer is a deliberately simple, dependency-free stand-in
// for the real GNN: a z-score threshold over each series' recent
// window. It exists so gn-sim and early integration tests can run
// end-to-end before the trained model is wired in.
type ReferenceScorer struct {
	Threshold float64 // z-score threshold, default ~3.0
}

func NewReferenceScorer() *ReferenceScorer {
	return &ReferenceScorer{Threshold: 3.0}
}

func (r *ReferenceScorer) Score(nodes []NodeFeatures, edges []EdgeFeatures) ([]Score, error) {
	out := make([]Score, 0, len(nodes)+len(edges))
	for _, n := range nodes {
		out = append(out, Score{ID: n.NodeID, Score: zScoreAnomaly(n.Window, r.Threshold)})
	}
	for _, e := range edges {
		out = append(out, Score{ID: e.EdgeID, Score: zScoreAnomaly(e.Window, r.Threshold)})
	}
	return out, nil
}

func zScoreAnomaly(window []float64, threshold float64) float64 {
	if len(window) < 2 {
		return 0
	}
	mean := 0.0
	for _, v := range window {
		mean += v
	}
	mean /= float64(len(window))

	variance := 0.0
	for _, v := range window {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(window))
	std := sqrt(variance)
	if std == 0 {
		return 0
	}
	last := window[len(window)-1]
	z := (last - mean) / std
	if z < 0 {
		z = -z
	}
	score := z / threshold
	if score > 1 {
		score = 1
	}
	return score
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	// Newton's method; avoids pulling in math just for this.
	z := x
	for i := 0; i < 20; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}

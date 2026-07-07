// Package gensynth is the "GenNets" synthetic generator: produces
// realistic synthetic topologies, telemetry, and labeled fault
// injections for GNN training and for the live-mode simulator.
package gensynth

import (
	"math"
	"math/rand"
)

type Archetype string

const (
	ArchetypeRing         Archetype = "ring"
	ArchetypeMesh         Archetype = "mesh"
	ArchetypeHierarchical Archetype = "hierarchical" // access/aggregation/core
	ArchetypeHybridQKD    Archetype = "hybrid_qkd"    // hierarchical + QKD overlay
)

type Config struct {
	Seed          int64
	NodeCount     int
	Archetype     Archetype
	QKDLinkRatio  float64 // fraction of edges that are QKD channels (hybrid_qkd only)
	FaultRate     float64 // faults injected per simulated hour
}

func DefaultConfig() Config {
	return Config{
		Seed:         42,
		NodeCount:    50,
		Archetype:    ArchetypeHierarchical,
		QKDLinkRatio: 0.1,
		FaultRate:    0.5,
	}
}

type SyntheticNode struct {
	ID   string
	Type string
}

type SyntheticEdge struct {
	ID    string
	Src   string
	Dst   string
	Type  string
}

type SyntheticTopology struct {
	Nodes []SyntheticNode
	Edges []SyntheticEdge
}

type FaultInjection struct {
	AtTick        int
	TargetNodeID  string
	TargetEdgeID  string
	Kind          string // "link_degradation" | "node_flap" | "qkd_drift" | "correlated_multi"
	GroundTruth   []string
}

// Generator produces a synthetic topology plus a schedule of fault
// injections; telemetry sample values are produced tick-by-tick via
// Tick() so gn-sim can stream them live onto NATS.
type Generator struct {
	cfg    Config
	rng    *rand.Rand
	topo   SyntheticTopology
	faults []FaultInjection
}

func New(cfg Config) *Generator {
	g := &Generator{cfg: cfg, rng: rand.New(rand.NewSource(cfg.Seed))}
	g.topo = g.buildTopology()
	g.faults = g.scheduleFaults()
	return g
}

func (g *Generator) Topology() SyntheticTopology { return g.topo }
func (g *Generator) Faults() []FaultInjection     { return g.faults }

func (g *Generator) buildTopology() SyntheticTopology {
	// Minimal hierarchical builder: one core, sqrt(n) aggregation nodes,
	// remainder as access/leaf nodes, ring-connected at each tier.
	// Real implementation should support all Archetype values; this is
	// a v0 reference sufficient to unblock gn-sim end-to-end.
	nodes := []SyntheticNode{{ID: "core-0", Type: "router"}}
	aggCount := int(math.Max(1, math.Sqrt(float64(g.cfg.NodeCount))))
	edges := []SyntheticEdge{}

	for i := 0; i < aggCount; i++ {
		aggID := idOf("agg", i)
		nodes = append(nodes, SyntheticNode{ID: aggID, Type: "router"})
		edges = append(edges, SyntheticEdge{ID: idOf("e-core-agg", i), Src: "core-0", Dst: aggID, Type: "fiber"})
	}

	remaining := g.cfg.NodeCount - 1 - aggCount
	for i := 0; i < remaining; i++ {
		leafID := idOf("leaf", i)
		parentAgg := idOf("agg", i%aggCount)
		leafType := "cell_site"
		edgeType := "microwave"
		if g.cfg.Archetype == ArchetypeHybridQKD && g.rng.Float64() < g.cfg.QKDLinkRatio {
			leafType = "qkd_node"
			edgeType = "qkd_channel"
		}
		nodes = append(nodes, SyntheticNode{ID: leafID, Type: leafType})
		edges = append(edges, SyntheticEdge{ID: idOf("e-agg-leaf", i), Src: parentAgg, Dst: leafID, Type: edgeType})
	}

	return SyntheticTopology{Nodes: nodes, Edges: edges}
}

func (g *Generator) scheduleFaults() []FaultInjection {
	if len(g.topo.Edges) == 0 {
		return nil
	}
	var out []FaultInjection
	// One fault roughly every 1/FaultRate ticks (tick == simulated minute).
	ticksPerFault := int(60.0 / math.Max(g.cfg.FaultRate, 0.01))
	for tick := ticksPerFault; tick < ticksPerFault*20; tick += ticksPerFault {
		e := g.topo.Edges[g.rng.Intn(len(g.topo.Edges))]
		kind := "link_degradation"
		if e.Type == "qkd_channel" {
			kind = "qkd_drift"
		}
		out = append(out, FaultInjection{
			AtTick:       tick,
			TargetEdgeID: e.ID,
			Kind:         kind,
			GroundTruth:  []string{e.ID, e.Src, e.Dst},
		})
	}
	return out
}

func idOf(prefix string, i int) string {
	return prefix + "-" + itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

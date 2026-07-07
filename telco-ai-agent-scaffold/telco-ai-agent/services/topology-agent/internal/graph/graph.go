// Package graph implements the in-memory network topology model:
// nodes (routers, OLT/ONT, cell sites, QKD nodes) and edges (fiber,
// microwave, QKD channel, logical links), with versioned deltas.
package graph

import "sync"

type NodeType int

const (
	NodeUnspecified NodeType = iota
	NodeRouter
	NodeOLT
	NodeONT
	NodeCellSite
	NodeQKD
)

type EdgeType int

const (
	EdgeUnspecified EdgeType = iota
	EdgeFiber
	EdgeMicrowave
	EdgeQKDChannel
	EdgeLogical
)

type Node struct {
	ID    string
	Type  NodeType
	Attrs map[string]string
}

type Edge struct {
	ID        string
	SrcNodeID string
	DstNodeID string
	Type      EdgeType
	Bandwidth float64
	LatencyMs float64
}

type Graph struct {
	mu      sync.RWMutex
	version int64
	nodes   map[string]Node
	edges   map[string]Edge
}

func New() *Graph {
	return &Graph{nodes: map[string]Node{}, edges: map[string]Edge{}}
}

func (g *Graph) Version() int64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.version
}

// UpsertNode adds/updates a node and bumps the graph version.
func (g *Graph) UpsertNode(n Node) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[n.ID] = n
	g.version++
}

// UpsertEdge adds/updates an edge and bumps the graph version.
func (g *Graph) UpsertEdge(e Edge) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.edges[e.ID] = e
	g.version++
}

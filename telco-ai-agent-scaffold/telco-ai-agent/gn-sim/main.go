// Command gn-sim is the full-stack simulator: it runs the GN engine's
// synthetic generator in "live mode", publishing synthetic topology
// and telemetry onto the real NATS subjects so the whole subagent
// pipeline (Topology -> Anomaly -> Fault Corr -> QKD Health ->
// Frontend) can be exercised end-to-end without real network data.
package main

import (
	"flag"
	"log"
	"time"

	"github.com/telco-ai-agent/gn-engine/gensynth"
)

func main() {
	nodeCount := flag.Int("nodes", 50, "number of synthetic nodes")
	archetype := flag.String("archetype", "hybrid_qkd", "ring|mesh|hierarchical|hybrid_qkd")
	natsURL := flag.String("nats", "nats://localhost:4222", "NATS JetStream URL")
	tickInterval := flag.Duration("tick", time.Second, "simulated tick interval")
	flag.Parse()

	cfg := gensynth.DefaultConfig()
	cfg.NodeCount = *nodeCount
	cfg.Archetype = gensynth.Archetype(*archetype)

	gen := gensynth.New(cfg)
	topo := gen.Topology()
	log.Printf("gn-sim: generated topology: %d nodes, %d edges (archetype=%s)",
		len(topo.Nodes), len(topo.Edges), cfg.Archetype)
	log.Printf("gn-sim: scheduled %d fault injections", len(gen.Faults()))
	log.Printf("gn-sim: would publish to %s every %s (TODO: NATS JetStream wiring)",
		*natsURL, *tickInterval)

	// TODO: connect to NATS, publish topology snapshot to
	// telco.topology.snapshot, then loop: emit per-tick telemetry to
	// telco.telemetry.raw.>, applying scheduled fault injections at
	// their tick, until interrupted.
}

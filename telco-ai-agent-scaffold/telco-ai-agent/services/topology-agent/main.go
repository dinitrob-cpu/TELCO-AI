// Command topology-agent maintains the live network graph: it
// consumes raw telemetry/discovery events from NATS, builds/updates
// the topology graph, and serves it over gRPC (TopologyService).
package main

import (
	"log"

	"github.com/telco-ai-agent/topology-agent/internal/graph"
)

func main() {
	g := graph.New()
	log.Printf("topology-agent: starting, graph version=%d", g.Version())
	// TODO: subscribe to telco.telemetry.raw.>, serve TopologyService on gRPC.
	select {}
}

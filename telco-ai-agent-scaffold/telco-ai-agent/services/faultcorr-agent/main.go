// Command faultcorr-agent subscribes to anomaly events and topology
// deltas, groups temporally + topologically adjacent anomalies into
// fault hypotheses, and serves them over gRPC (FaultCorrService).
package main

import (
	"log"
	"time"

	"github.com/telco-ai-agent/faultcorr-agent/internal/correlate"
)

func main() {
	corr := correlate.New(5 * time.Minute)
	log.Println("faultcorr-agent: starting, correlation window =", corr.Window())
	// TODO: subscribe to telco.anomaly.events + telco.topology.delta,
	// serve FaultCorrService on gRPC.
	select {}
}

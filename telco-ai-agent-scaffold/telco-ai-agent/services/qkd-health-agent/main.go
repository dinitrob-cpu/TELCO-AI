// Command qkd-health-agent monitors quantum key distribution link
// health (QBER, key rate, sifting ratio, channel drift) and raises
// anomaly-compatible events so Fault Corr needs no QKD-specific logic.
package main

import (
	"log"

	"github.com/telco-ai-agent/qkd-health-agent/internal/health"
)

func main() {
	monitor := health.NewMonitor()
	log.Println("qkd-health-agent: starting")
	_ = monitor
	// TODO: subscribe to raw QKD link telemetry, publish telco.qkd.health
	// and telco.anomaly.events, serve QKDHealthService on gRPC.
	select {}
}

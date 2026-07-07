// Command orchestrator is the entrypoint for the TELCO AI Agent Orchestrator.
// It owns the frontend-facing gRPC gateway, the subagent registry, and
// saga-style coordination between subagents over NATS JetStream.
package main

import (
	"log"

	"github.com/telco-ai-agent/orchestrator/internal/gwserver"
	"github.com/telco-ai-agent/orchestrator/internal/registry"
)

func main() {
	reg := registry.New()
	srv := gwserver.New(reg)

	log.Println("orchestrator: starting gRPC gateway on :8080, NATS at nats://localhost:4222")
	if err := srv.ListenAndServe(":8080"); err != nil {
		log.Fatalf("orchestrator: fatal: %v", err)
	}
}

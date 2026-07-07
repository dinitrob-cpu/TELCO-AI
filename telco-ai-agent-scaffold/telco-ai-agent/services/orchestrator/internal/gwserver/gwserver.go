// Package gwserver hosts the gRPC gateway exposed to the Next.js
// frontend (via grpc-gateway/Connect, gRPC-Web compatible).
package gwserver

import (
	"net"

	"google.golang.org/grpc"

	"github.com/telco-ai-agent/orchestrator/internal/registry"
)

type Server struct {
	grpcServer *grpc.Server
	registry   *registry.Registry
}

func New(reg *registry.Registry) *Server {
	return &Server{
		grpcServer: grpc.NewServer(),
		registry:   reg,
	}
}

func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// TODO: register composed gateway services (topology, anomaly,
	// faultcorr, qkd) as they come online.
	return s.grpcServer.Serve(lis)
}

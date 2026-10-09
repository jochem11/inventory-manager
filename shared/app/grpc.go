package app

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/jochem11/inventory-manager/shared/env"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/stats"
)

// ServeGRPC serves gRPC on GRPC_ADDR (default defaultAddr) until the service
// stops, then stops gracefully: running calls finish first. register adds the
// service's handlers. The server also has:
//   - a span per call, continuing the caller's trace (health checks left out)
//   - the health service, reporting service as serving, for Kubernetes probes
//   - reflection, so grpcurl works without the .proto files
func (a *App) ServeGRPC(defaultAddr, service string, register func(s *grpc.Server)) error {
	server := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler(
		otelgrpc.WithFilter(func(info *stats.RPCTagInfo) bool {
			return !strings.HasPrefix(info.FullMethodName, "/grpc.health.v1.Health/")
		}),
	)))
	register(server)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)
	reflection.Register(server)

	addr := env.Get("GRPC_ADDR", defaultAddr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	go func() {
		<-a.ctx.Done()
		log.Println("shutting down")
		healthServer.Shutdown()
		server.GracefulStop()
	}()

	log.Printf("gRPC server listening on %s", lis.Addr())
	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

package server

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/internal/snapshot"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	secretService "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

func NewServer(ctx context.Context, config cache.Cache, cb server.Callbacks) server.Server {
	return server.NewServer(ctx, config, cb)
}

func Run(socketPath string, l log.Logger) (net.Listener, *grpc.Server, error) {
	// Create a cache
	c := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Create the snapshot that we'll serve to Envoy
	s := snapshot.GenerateSnapshot()
	if err := s.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", s, err)
		return nil, nil, err
	}
	l.Debugf("will serve snapshot %+v", s)

	// Add the snapshot to the cache
	if err := c.SetSnapshot(context.Background(), "", s); err != nil {
		l.Errorf("snapshot error %q for %+v", err, s)
		return nil, nil, err
	}

	// Run the xDS server
	ctx := context.Background()
	cb := &test.Callbacks{
		Debug: l.Debug,
	}
	srv := server.NewServer(ctx, c, cb)

	lis, grpcServer, err := runServer(srv, socketPath, l)
	if err != nil {
		return nil, nil, err
	}

	return lis, grpcServer, nil
}

// runServer starts an xDS server at the given port.
func runServer(srv server.Server, sockerPath string, l log.Logger) (net.Listener, *grpc.Server, error) {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems. Keepalive timeouts based on connection_keepalive parameter https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/examples#dynamic
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("unix", sockerPath)
	if err != nil {
		l.Errorf("%+v", err)
		return nil, nil, err
	}

	registerServer(grpcServer, srv)

	go func() {
		l.Infof("sds server listening on socket: %s\n", sockerPath)
		if err = grpcServer.Serve(lis); err != nil {
			l.Errorf("%+v", err)
		}
	}()

	return lis, grpcServer, nil
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	// register services
	secretService.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	grpcHealth.RegisterHealthServer(grpcServer, health.NewServer())
}

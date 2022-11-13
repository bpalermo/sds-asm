package server

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/internal/snapshot"
	"github.com/bpalermo/sds-asm/pkg/subscription"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"os"
	"os/signal"
	"syscall"
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

type SdsServer struct {
	l     log.Logger
	c     cache.SnapshotCache
	sigCh chan os.Signal
	srv   *grpc.Server
	s     *subscription.Subscriber
}

func New(awsRegion string, awsEndpoint string, l log.Logger) (*SdsServer, error) {
	subscriber, err := newSubscriber(awsRegion, awsEndpoint, l)
	if err != nil {
		return nil, err
	}

	return &SdsServer{
		l:     l,
		c:     newSnapshotCache(l),
		sigCh: newSigCh(),
		srv:   newGrpcServer(),
		s:     subscriber,
	}, nil
}

func newSnapshotCache(l log.Logger) cache.SnapshotCache {
	// Create a cache
	c := cache.NewSnapshotCache(false, cache.IDHash{}, l)
	newSnapshot(c, l)
	return c
}

func newSnapshot(c cache.SnapshotCache, l log.Logger) {
	// Create the snapshot that we'll serve to Envoy
	s := snapshot.GenerateSnapshot()
	if err := s.Consistent(); err != nil {
		l.Fatal().Err(err).Interface("snapshot", s).Msg("snapshot inconsistency")
	}
	l.Debug().Interface("snapshot", s).Msg("will serve snapshot")

	// Add the snapshot to the cache
	if err := c.SetSnapshot(context.Background(), "test-01", s); err != nil {
		l.Fatal().Err(err).Interface("snapshot", s).Msg("snapshot error")
	}
}

func newGrpcServer() *grpc.Server {
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

	return grpc.NewServer(grpcOptions...)
}

func newSigCh() chan os.Signal {
	sigCh := make(chan os.Signal, 1)

	// register for interrupt (Ctrl+C) and SIGTERM (docker)
	signal.Notify(sigCh,
		os.Interrupt,
		syscall.SIGTERM,
	)

	return sigCh
}

func newSubscriber(awsRegion string, awsEndpoint string, l log.Logger) (*subscription.Subscriber, error) {
	return subscription.New(awsRegion, awsEndpoint, l)
}

func (s *SdsServer) Run(socketPath string) error {
	ctx := context.Background()

	go func() {
		err := s.startServer(ctx, socketPath)
		if err != nil {
			s.l.Fatal().Err(err).Msg("could not start server")
		}
	}()

	for {
		sig := <-s.sigCh
		// stop signal
		s.l.Infof("got signal %v, attempting graceful shutdown", sig)
		s.s.Stop()
		s.srv.GracefulStop()
		close(s.sigCh)
		return nil
	}
}

func (s *SdsServer) startServer(ctx context.Context, socketPath string) error {
	srv := server.NewServer(ctx, s.c, &Callbacks{})

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		s.l.Error().Err(err).Msg("could not listen")
		return err
	}
	s.l.Debug().Str("path", socketPath).Msg("listening at socket")

	registerServer(s.srv, srv)

	if err = s.srv.Serve(lis); err != nil {
		s.l.Error().Err(err).Msg("could not serve")
	}

	s.l.Info().Msg("clean shutdown")
	return nil
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	// register services
	secretService.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	grpcHealth.RegisterHealthServer(grpcServer, health.NewServer())
}

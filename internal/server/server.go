package server

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/internal/snapshot"
	"github.com/bpalermo/sds-asm/pkg/subscription"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/grpc/health"
	grpcHealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/anypb"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	zLog "github.com/rs/zerolog/log"

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
	cache      cache.SnapshotCache
	callbacks  *Callbacks
	grpcServer *grpc.Server
	l          log.Logger
	notifyCh   chan *tlsV3.Secret
	sds        server.Server
	sigCh      chan os.Signal
}

func NewServer(awsRegion string, awsEndpoint string, l log.Logger) (*SdsServer, error) {
	subscriber, err := newSubscriber(awsRegion, awsEndpoint, l)
	if err != nil {
		return nil, err
	}

	grpcServer := newGrpcServer()
	snapshotCache := newSnapshotCache(l)
	callbacks := NewCallbacks(subscriber, l)

	return &SdsServer{
		cache:      snapshotCache,
		callbacks:  callbacks,
		grpcServer: grpcServer,
		l:          l,
		notifyCh:   make(chan *tlsV3.Secret),
		sigCh:      newSigCh(),
		sds:        newSdsServer(context.Background(), grpcServer, snapshotCache, callbacks),
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
	s, err := snapshot.GenerateSnapshot()
	if err != nil {
		zLog.Error().Err(err).Interface("snapshot", s).Msg("could not generate snapshot")
		return
	}
	if err = s.Consistent(); err != nil {
		zLog.Error().Err(err).Interface("snapshot", s).Msg("snapshot inconsistency")
		return
	}
	zLog.Debug().Interface("snapshot", s).Msg("serving snapshot")

	// Add the snapshot to the cache
	if err := c.SetSnapshot(context.Background(), "test-01", s); err != nil {
		l.Errorf("snapshot error %+v: %+v", s, err)
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

func newSdsServer(ctx context.Context, grpcServer *grpc.Server, cache cache.SnapshotCache, callbacks *Callbacks) server.Server {
	srv := server.NewServer(ctx, cache, callbacks)
	registerServer(grpcServer, srv)
	return srv
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
	return subscription.NewSubscriber(awsRegion, awsEndpoint, l)
}

func (s *SdsServer) Run(socketPath string) error {
	go func() {
		err := s.startServer(socketPath)
		if err != nil {
			s.l.Errorf("could not start server %+v", err)
		}
	}()

	for {
		select {
		case secret := <-s.notifyCh:
			tlsSecret, _ := anypb.New(secret)
			snap, err := cache.NewSnapshot("2", map[resource.Type][]types.Resource{
				resource.SecretType: {
					tlsSecret,
				},
			})
			if err != nil {
				s.l.Errorf("could not create snapshot %+v", err)
			}
			err = s.cache.SetSnapshot(context.Background(), "test-01", snap)
			if err != nil {
				s.l.Errorf("could not set snapshot %+v", err)
			}
		case sig := <-s.sigCh:
			// stop signal
			s.l.Infof("got signal %v, attempting graceful shutdown", sig)
			s.callbacks.Stop()
			s.grpcServer.GracefulStop()
			close(s.sigCh)
			return nil
		}
	}
}

func (s *SdsServer) startServer(socketPath string) error {
	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		s.l.Errorf("could not listen %+v", err)
		return err
	}
	s.l.Debugf("listening at socket %s", socketPath)

	if err = s.grpcServer.Serve(lis); err != nil {
		s.l.Errorf("could not serve %+v", err)
	}

	s.l.Infof("clean shutdown")
	return nil
}

func registerServer(grpcServer *grpc.Server, server server.Server) {
	// register services
	secretService.RegisterSecretDiscoveryServiceServer(grpcServer, server)
	grpcHealth.RegisterHealthServer(grpcServer, health.NewServer())
}

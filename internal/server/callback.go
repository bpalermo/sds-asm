package server

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/pkg/subscription"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	zLog "github.com/rs/zerolog/log"
	"sync"
)

type Callbacks struct {
	Logger         log.Logger
	Signal         chan struct{}
	Fetches        int
	Requests       int
	DeltaRequests  int
	DeltaResponses int
	subscriber     subscription.SubscribeAPI
	mu             sync.Mutex
}

func NewCallbacks(s *subscription.Subscriber, l log.Logger) *Callbacks {
	return &Callbacks{
		Logger:     l,
		subscriber: s,
	}
}

func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Logger.Debugf("startServer callbacks fetches=%d requests=%d", cb.Fetches, cb.Requests)
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	zLog.Debug().Str("type", typ).Int64("id", id).Msg("stream closed")
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64) {
	zLog.Debug().Int64("id", id).Msg("stream closed")
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	cb.Logger.Debugf("delta stream %d open, type %s", id, typ)
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64) {
	cb.Logger.Debugf("delta stream %d closed", id)
}

func (cb *Callbacks) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
	zLog.Debug().Interface("request", req).Msg("OnStreamRequest")
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.Requests++
	if req.TypeUrl == resource.SecretType {
		for _, name := range req.ResourceNames {
			cb.Logger.Debugf("stream resource %s", name)
			cb.subscriber.Subscribe(req.Node.Id, name)
		}
	}

	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnStreamResponse(context.Context, int64, *discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
}

func (cb *Callbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
	cb.Logger.Debugf("delta stream request id %d, request %+v, response %+v", id, req, res)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaResponses++
}
func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	cb.Logger.Debugf("delta stream request id %d, request %+v", id, req)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaRequests++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}

	return nil
}

func (cb *Callbacks) OnFetchRequest(_ context.Context, req *discovery.DiscoveryRequest) error {
	cb.Logger.Debugf("on fetch request %+v", req)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Fetches++
	if cb.Signal != nil {
		close(cb.Signal)
		cb.Signal = nil
	}
	return nil
}

func (cb *Callbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {}

func (cb *Callbacks) Stop() {
	cb.subscriber.Stop()
}

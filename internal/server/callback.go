package server

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/pkg/subscription"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
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

func (cb *Callbacks) Report() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Logger.Debug().Msgf("startServer callbacks fetches=%d requests=%d\n", cb.Fetches, cb.Requests)
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	cb.Logger.Debug().Str("type", typ).Msgf("stream %d open", id)
	return nil
}

func (cb *Callbacks) OnStreamClosed(id int64) {
	cb.Logger.Debug().Msgf("stream %d closed\n", id)
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	cb.Logger.Debug().Str("type", typ).Msgf("delta stream %d open", id)
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(id int64) {
	cb.Logger.Debug().Msgf("delta stream %d closed\n", id)
}

func (cb *Callbacks) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
	cb.Logger.Debug().Interface("request", req).Msgf("stream request id %d", id)
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.Requests++
	if len(req.ResourceNames) > 0 {
		cb.subscriber.Subscribe(req.ResourceNames[0])
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
	cb.Logger.Debug().Interface("request", req).Interface("response", res).Msgf("delta stream request id %d", id)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.DeltaResponses++
}
func (cb *Callbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	cb.Logger.Debug().Interface("request", req).Msgf("delta stream request id %d", id)
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
	cb.Logger.Debug().Interface("request", req).Msgf("on fetch request")
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

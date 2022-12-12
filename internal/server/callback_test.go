package server

import (
	"context"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockSubscribe struct {
}

func (s *MockSubscribe) Subscribe(_ string, _ string) {
}

func (s *MockSubscribe) Stop() {
}

func TestCallbacks_OnDeltaStreamClosed(t *testing.T) {
	c := &Callbacks{}
	c.OnDeltaStreamClosed(1)
}

func TestCallbacks_OnDeltaStreamOpen(t *testing.T) {
	c := &Callbacks{}
	err := c.OnDeltaStreamOpen(context.Background(), 1, "")
	assert.Nil(t, err)
}

func TestCallbacks_OnFetchRequest(t *testing.T) {
	c := &Callbacks{}
	err := c.OnFetchRequest(context.Background(), &discovery.DiscoveryRequest{})
	assert.Nil(t, err)
}

func TestCallbacks_OnFetchResponse(t *testing.T) {
	c := &Callbacks{}
	c.OnFetchResponse(&discovery.DiscoveryRequest{}, &discovery.DiscoveryResponse{})
}

func TestCallbacks_Report(t *testing.T) {
	c := &Callbacks{}
	c.Report()
}

func TestCallbacks_OnStreamRequest(t *testing.T) {
	c := &Callbacks{
		subscriber: &MockSubscribe{},
	}
	err := c.OnStreamRequest(1, &discovery.DiscoveryRequest{
		Node: &corev3.Node{
			Id: "test-01",
		},
		TypeUrl: resource.SecretType,
		ResourceNames: []string{
			"test",
		},
	})

	assert.Nil(t, err)
}

func TestCallbacks_OnStreamClose(t *testing.T) {
	c := &Callbacks{}
	c.OnStreamClosed(1)
}

func TestCallbacks_OnStreamDeltaRequest(t *testing.T) {
	c := &Callbacks{}
	err := c.OnStreamDeltaRequest(1, &discovery.DeltaDiscoveryRequest{})
	assert.Nil(t, err)
}

func TestCallbacks_OnStreamDeltaResponse(t *testing.T) {
	c := &Callbacks{}
	c.OnStreamDeltaResponse(1, &discovery.DeltaDiscoveryRequest{}, &discovery.DeltaDiscoveryResponse{})
}

func TestCallbacks_OnStreamOpen(t *testing.T) {
	c := &Callbacks{}
	err := c.OnStreamOpen(context.Background(), 1, "")
	assert.Nil(t, err)
}

func TestCallbacks_OnStreamResponse(t *testing.T) {

}

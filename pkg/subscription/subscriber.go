package subscription

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bpalermo/sds-asm/internal/aws"
	"github.com/bpalermo/sds-asm/internal/log"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"sync"
	"time"
)

type SubscribeAPI interface {
	Subscribe(nodeId string, secretId string)
	Stop()
}

type Subscriber struct {
	api           aws.SecretsManagerAPI
	l             log.Logger
	maxInterval   time.Duration
	minInterval   time.Duration
	mu            sync.Mutex
	notifyCh      chan *tlsV3.Secret
	subscriptions map[*string]*Watcher
}

func NewSubscriber(awsRegion string, awsEndpoint string, l log.Logger) (*Subscriber, error) {
	const (
		defaultMinInterval = 25 * time.Second
		defaultMaxInterval = 35 * time.Second
	)

	cfg, err := aws.LoadConfig(context.Background(), awsEndpoint, awsRegion, l)
	if err != nil {
		return nil, err
	}

	s := &Subscriber{
		api:           secretsmanager.NewFromConfig(cfg),
		l:             l,
		minInterval:   defaultMinInterval,
		maxInterval:   defaultMaxInterval,
		notifyCh:      make(chan *tlsV3.Secret),
		subscriptions: make(map[*string]*Watcher, 0),
	}

	return s, nil
}

func (s *Subscriber) Stop() {
	for _, w := range s.subscriptions {
		w.Stop()
	}
}

func (s *Subscriber) Subscribe(nodeId string, secretId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.subscriptions[&secretId]; !found {
		// subscribe
		w := NewWatcher(&nodeId, &secretId, s.api, s.notifyCh, s.l)
		w.Start()

		s.subscriptions[&secretId] = w
	}
}

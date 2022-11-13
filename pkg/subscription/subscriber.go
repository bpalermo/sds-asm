package subscription

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bpalermo/sds-asm/internal/aws"
	"github.com/bpalermo/sds-asm/internal/log"
	"time"
)

type SubscribeAPI interface {
	Subscribe(secretId string)
}

type Subscriber struct {
	api           aws.SecretsManagerAPI
	l             log.Logger
	maxInterval   time.Duration
	minInterval   time.Duration
	subscriptions map[*string]*Watcher
}

func New(awsRegion string, awsEndpoint string, l log.Logger) (*Subscriber, error) {
	const (
		defaultMinInterval = 25 * time.Second
		defaultMaxInterval = 35 * time.Second
	)

	cfg, err := aws.LoadConfig(context.Background(), awsEndpoint, awsRegion)
	if err != nil {
		return nil, err
	}

	s := &Subscriber{
		api:           secretsmanager.NewFromConfig(cfg),
		l:             l,
		minInterval:   defaultMinInterval,
		maxInterval:   defaultMaxInterval,
		subscriptions: make(map[*string]*Watcher, 0),
	}

	return s, nil
}

func (s *Subscriber) Stop() {
	for _, w := range s.subscriptions {
		w.Stop()
	}
}

func (s *Subscriber) Subscribe(secretId string) {
	if _, found := s.subscriptions[&secretId]; !found {
		// subscribe
		w := NewWatcher(nil, &secretId, s.l)
		w.Start()

		s.subscriptions[&secretId] = w
	}
}

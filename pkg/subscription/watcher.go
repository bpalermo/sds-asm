package subscription

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bpalermo/sds-asm/internal/aws"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/pkg/ticker"
	"os"
	"time"
)

type WatchAPI interface {
	Start()
}

type WatchOption func(watcher *Watcher)

type Watcher struct {
	l                      log.Logger
	currentSecretVersionId *string
	notifyCh               chan []byte
	signalCh               chan os.Signal
	secretId               *string
	ticker                 *ticker.RandomTicker
	api                    aws.SecretsManagerAPI
}

func NewWatcher(notifyCh chan []byte, secretId *string, l log.Logger, opts ...WatchOption) *Watcher {
	const (
		defaultMinInterval = 25 * time.Second
		defaultMaxInterval = 35 * time.Second
	)

	w := &Watcher{
		l:                      l,
		currentSecretVersionId: nil,
		notifyCh:               notifyCh,
		signalCh:               make(chan os.Signal),
		secretId:               secretId,
		ticker:                 ticker.NewRandomTicker(defaultMinInterval, defaultMaxInterval),
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(w)
	}

	return w

}

func WithApi(api aws.SecretsManagerAPI) WatchOption {
	return func(w *Watcher) {
		w.api = api
	}
}

func WithInterval(min, max time.Duration) WatchOption {
	return func(w *Watcher) {
		w.ticker = ticker.NewRandomTicker(min, max)
	}
}

func (w *Watcher) Start() {
	w.ticker.Start()

	for {
		select {
		case <-w.ticker.C:
			w.l.Info().Str("secret", *w.secretId).Msg("tick")
			secret, err := w.checkSecret()
			if err != nil {
				w.l.Err(err).Msg("failed to fetch secret")
			}

			if secret != nil {
				w.l.Info().Msg("new secret")
				w.notifyCh <- secret
			}
		case s := <-w.signalCh:
			// stop signal
			w.l.Infof("got signal %v, attempting graceful shutdown", s)
			w.ticker.Stop()
		}
	}
}

func (w *Watcher) Stop() {
	w.ticker.Stop()
}

func (w *Watcher) checkSecret() ([]byte, error) {
	if w.currentSecretVersionId == nil {
		versionId, secret, err := w.fetchSecret()

		w.currentSecretVersionId = versionId
		return secret, err
	}

	return nil, nil
}

func (w *Watcher) fetchSecret() (*string, []byte, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: w.secretId,
	}
	response, err := w.api.GetSecretValue(context.Background(), input, nil)
	if err != nil {
		return nil, nil, err
	}

	return response.VersionId, response.SecretBinary, nil
}

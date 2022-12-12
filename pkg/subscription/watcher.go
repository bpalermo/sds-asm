package subscription

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bpalermo/sds-asm/internal/aws"
	"github.com/bpalermo/sds-asm/internal/helper"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/pkg/secret"
	"github.com/bpalermo/sds-asm/pkg/ticker"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	zLog "github.com/rs/zerolog/log"
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
	nodeId                 *string
	notifyCh               chan *tlsV3.Secret
	signalCh               chan os.Signal
	secretId               *string
	ticker                 *ticker.RandomTicker
	api                    aws.SecretsManagerAPI
}

func NewWatcher(nodeId *string, secretId *string, api aws.SecretsManagerAPI, notifyCh chan *tlsV3.Secret, l log.Logger, opts ...WatchOption) *Watcher {
	const (
		defaultMinInterval = 25 * time.Second
		defaultMaxInterval = 35 * time.Second
	)

	w := &Watcher{
		api:                    api,
		currentSecretVersionId: nil,
		l:                      l,
		nodeId:                 nodeId,
		notifyCh:               notifyCh,
		secretId:               secretId,
		signalCh:               make(chan os.Signal),
		ticker:                 ticker.NewRandomTicker(defaultMinInterval, defaultMaxInterval),
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(w)
	}

	return w

}

func WithInterval(min, max time.Duration) WatchOption {
	return func(w *Watcher) {
		w.ticker = ticker.NewRandomTicker(min, max)
	}
}

func (w *Watcher) Start() {
	// initial check
	w.check()
	w.ticker.Start()

	for {
		select {
		case <-w.ticker.C:
			w.l.Debugf("tick for secret %s", *w.secretId)
			w.check()
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

func (w *Watcher) check() {
	zLog.Debug().
		Str("nodeId", *w.nodeId).
		Str("secretId", *w.secretId).
		Msg("failed to fetch secret")
	certChain, privateKey, err := w.checkSecret()
	if err != nil {
		zLog.Error().Err(err).Msg("failed to fetch secret")
		return
	}

	notify(*w.secretId, certChain, privateKey, w.notifyCh)
}

func (w *Watcher) checkSecret() ([]byte, []byte, error) {
	if w.currentSecretVersionId == nil {
		versionId, certChain, privateKey, err := w.fetchSecret()

		w.currentSecretVersionId = versionId
		return certChain, privateKey, err
	}

	return nil, nil, nil
}

func (w *Watcher) fetchSecret() (versionId *string, certChain []byte, privateKey []byte, err error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: w.secretId,
	}
	response, err := w.api.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, nil, nil, err
	}

	certChain, privateKey, err = parseSecret(response)
	if err != nil {
		return nil, nil, nil, err
	}

	return response.VersionId, certChain, privateKey, nil
}

func parseSecret(getSecretOutput *secretsmanager.GetSecretValueOutput) ([]byte, []byte, error) {
	secretsManagerSecret, err := secret.Unmarshal(getSecretOutput.SecretString)
	if err != nil {
		return nil, nil, err
	}
	return secretsManagerSecret.CertificateChain, secretsManagerSecret.PrivateKey, nil
}

func notify(name string, certChain []byte, privateKey []byte, ch chan *tlsV3.Secret) {
	if ch == nil {
		zLog.Warn().Msg("watcher notify channel is not set")
		return
	}

	ch <- helper.TlsSecretFromBytes(name, certChain, privateKey)
}

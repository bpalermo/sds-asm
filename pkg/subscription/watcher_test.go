package subscription

import (
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
)

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	expectedSecretId = "test"
)

var (
	expectedTlsSecret = &tlsV3.Secret{
		Name: expectedSecretId,
		Type: &tlsV3.Secret_TlsCertificate{
			TlsCertificate: &tlsV3.TlsCertificate{
				CertificateChain: &v3.DataSource{
					Specifier: &v3.DataSource_InlineBytes{
						InlineBytes: make([]uint8, 0),
					},
				},
				PrivateKey: &v3.DataSource{
					Specifier: &v3.DataSource_InlineBytes{
						InlineBytes: make([]uint8, 0),
					},
				},
			},
		},
	}
)

type mockSecretsManagerAPI struct {
}

func (m mockSecretsManagerAPI) GetSecretValue(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	arn := "dummy-arn"
	createDate := time.Now().Add(-time.Hour * 12) // 12 hours ago
	versionId := "very-random-uuid"
	otherVersionId := "other-random-uuid"
	versionStages := []string{"hello", "versionStage-42", "AWSCURRENT"}
	otherVersionStages := []string{"AWSPREVIOUS"}
	versionIdsToStages := make(map[string][]string)
	versionIdsToStages[versionId] = versionStages
	versionIdsToStages[otherVersionId] = otherVersionStages
	secretId := "dummy-secret-name"
	secretString := `{"privateKey": "", "certificateChain": ""}`

	return &secretsmanager.GetSecretValueOutput{
		ARN:           &arn,
		CreatedDate:   &createDate,
		Name:          &secretId,
		SecretString:  &secretString,
		VersionId:     &versionId,
		VersionStages: versionStages,
	}, nil
}

func TestNewWatcher(t *testing.T) {
	notifyCh := make(chan *tlsV3.Secret)
	w := newWatcher(notifyCh)

	assert.Equal(t, expectedSecretId, *w.secretId)
	assert.NotNil(t, w.ticker)
}

func TestWatcher_Start(t *testing.T) {
	notifyCh := make(chan *tlsV3.Secret)
	w := newWatcher(notifyCh)
	go w.Start()

	s := <-notifyCh
	assert.Equal(t, expectedTlsSecret, s)

	w.signalCh <- os.Interrupt
}

func newWatcher(notifyCh chan *tlsV3.Secret) *Watcher {
	l := log.Logger{}
	nodeId := "test-01"
	secretId := "test"

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return NewWatcher(&nodeId, &secretId, mockSecretsManagerAPI{}, notifyCh, l, WithInterval(50*time.Millisecond, 100*time.Millisecond))
}

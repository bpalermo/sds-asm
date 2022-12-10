package subscription

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
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
	secretBinary := []byte{128, 56, 44, 123}

	return &secretsmanager.GetSecretValueOutput{
		ARN:           &arn,
		CreatedDate:   &createDate,
		Name:          &secretId,
		SecretBinary:  secretBinary,
		VersionId:     &versionId,
		VersionStages: versionStages,
	}, nil
}

func TestNewWatcher(t *testing.T) {
	notifyCh := make(chan []byte)
	w := newWatcher(notifyCh)

	secretId := "test"

	assert.Equal(t, &secretId, w.secretId)
	assert.NotNil(t, w.ticker)
}

func TestWatcher_Start(t *testing.T) {
	notifyCh := make(chan []byte)
	w := newWatcher(notifyCh)
	go w.Start()

	s := <-notifyCh
	assert.Equal(t, []byte{128, 56, 44, 123}, s)

	w.signalCh <- os.Interrupt
}

func newWatcher(notifyCh chan []byte) *Watcher {
	l := log.Logger{}
	secretId := "test"

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return NewWatcher(notifyCh, &secretId, l, WithApi(mockSecretsManagerAPI{}), WithInterval(50*time.Millisecond, 100*time.Millisecond))
}

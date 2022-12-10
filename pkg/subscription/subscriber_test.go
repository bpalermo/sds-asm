package subscription

import (
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSubscriber(t *testing.T) {
	l := log.Logger{}
	s, err := New("", "", l)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, s.api)
	assert.NotNil(t, s.subscriptions)

	s.Stop()
}

package server

import (
	"context"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(context.Background(), nil, nil)
	assert.NotNil(t, s)
}

func TestRun(t *testing.T) {
	l := log.Logger{}
	lis, srv, err := Run("/tmp/api.sock", l)

	defer srv.Stop()
	defer func(lis net.Listener) {
		_ = lis.Close()
	}(lis)

	assert.Nil(t, err)
}

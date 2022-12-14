package server

import (
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"syscall"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	l := log.Logger{}
	s, err := NewServer("", "", l)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, s.cache)
	assert.NotNil(t, s.sigCh)
	assert.NotNil(t, s.grpcServer)
	assert.NotNil(t, s.callbacks)
}

func TestSdsServer_Run(t *testing.T) {
	l := log.Logger{}
	s, err := NewServer("", "", l)
	assert.Nil(t, err)

	errs := &errgroup.Group{}

	errs.Go(func() error {
		err := s.Run("/tmp/sock.api")
		if err != nil {
			return err
		}
		return nil
	})

	time.Sleep(1 * time.Second)
	s.sigCh <- syscall.SIGTERM

	err = errs.Wait()
	assert.Nil(t, err)
}

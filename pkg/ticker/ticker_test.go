package ticker

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"testing"
	"time"
)

func TestNewRandomTicker(t *testing.T) {
	ticker := NewRandomTicker(50*time.Microsecond, 75*time.Microsecond)
	assert.NotNil(t, ticker)
}

func TestRandomTicker_Start(t *testing.T) {
	ticker := NewRandomTicker(50*time.Microsecond, 75*time.Microsecond)
	ticker.Start()
	time.Sleep(1 * time.Second)
	ticker.Stop()
}

func TestRandomTicker_loop(t *testing.T) {
	ticker := NewRandomTicker(50*time.Microsecond, 75*time.Microsecond)

	errs := &errgroup.Group{}

	errs.Go(func() error {
		return ticker.loop()
	})

	time.Sleep(1 * time.Second)
	ticker.Stop()

	err := errs.Wait()
	assert.Nil(t, err)
}

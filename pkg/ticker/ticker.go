package ticker

import (
	"log"
	"math/rand"
	"time"
)

// RandomTicker is similar to time.Ticker but ticks at random intervals between
// the min and max duration values (stored internally as int64 nanosecond
// counts).
type RandomTicker struct {
	C      chan time.Time
	stopCh chan chan struct{}
	min    int64
	max    int64
}

// NewRandomTicker returns a pointer to an initialized instance of the
// RandomTicker. Min and max are durations of the shortest and longest allowed
// ticks. Ticker will run in a goroutine until explicitly stopped.
func NewRandomTicker(min, max time.Duration) *RandomTicker {
	return &RandomTicker{
		C:      make(chan time.Time),
		stopCh: make(chan chan struct{}),
		min:    min.Nanoseconds(),
		max:    max.Nanoseconds(),
	}
}

// Start initiates the ticker goroutine.
func (rt *RandomTicker) Start() {
	go func() {
		err := rt.loop()
		if err != nil {
			log.Fatal(err)
		}
	}()
}

// Stop terminates the ticker goroutine and closes the C channel.
func (rt *RandomTicker) Stop() {
	c := make(chan struct{})
	rt.stopCh <- c
	<-c
}

func (rt *RandomTicker) loop() error {
	defer close(rt.C)
	t := time.NewTimer(rt.nextInterval())
	for {
		// either a stop signal or a timeout
		select {
		case c := <-rt.stopCh:
			t.Stop()
			close(c)
			return nil
		case <-t.C:
			select {
			case rt.C <- time.Now():
				t.Stop()
				t = time.NewTimer(rt.nextInterval())
			default:
				// there could be none receiving...
			}
		}
	}
}

func (rt *RandomTicker) nextInterval() time.Duration {
	interval := rand.Int63n(rt.max-rt.min) + rt.min
	return time.Duration(interval) * time.Nanosecond
}

package log

import (
	"github.com/rs/zerolog"
	"testing"
)

func TestLogger_Debugf(t *testing.T) {
	l := Logger{}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	l.Debugf("works")
}

func TestLogger_Errorf(t *testing.T) {
	l := Logger{}
	l.Errorf("works %s", "yeah")
}

func TestLogger_Infof(t *testing.T) {
	l := Logger{}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	l.Infof("works %s", "yeah")
}

func TestLogger_Warnf(t *testing.T) {
	l := Logger{}
	l.Warnf("works %s", "yeah")
}

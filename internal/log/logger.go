package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	zerolog.Logger
}

// Debugf log to stdout only if Debug is true.
func (logger Logger) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Infof log to stdout only if Debug is true.
func (logger Logger) Infof(format string, args ...interface{}) {
	log.Info().Msgf(format, args...)
}

// Warnf log to stdout always.
func (logger Logger) Warnf(format string, args ...interface{}) {
	log.Warn().Msgf(format, args...)
}

// Errorf log to stdout always.
func (logger Logger) Errorf(format string, args ...interface{}) {
	log.Error().Msgf(format, args...)
}

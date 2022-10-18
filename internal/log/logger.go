package log

import "github.com/rs/zerolog/log"

type Logger struct {
	Debug bool
}

// Debugf log to stdout only if Debug is true.
func (logger Logger) Debugf(format string, args ...interface{}) {
	if logger.Debug {
		log.Printf(format+"\n", args...)
	}
}

// Infof log to stdout only if Debug is true.
func (logger Logger) Infof(format string, args ...interface{}) {
	if logger.Debug {
		log.Printf(format+"\n", args...)
	}
}

// Warnf log to stdout always.
func (logger Logger) Warnf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

// Errorf log to stdout always.
func (logger Logger) Errorf(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

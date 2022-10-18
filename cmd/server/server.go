package main

import (
	"flag"
	"github.com/bpalermo/sds-asm/internal/log"
	"github.com/bpalermo/sds-asm/internal/server"
	"github.com/rs/zerolog"
	"os"
)

var (
	l          log.Logger
	debug      bool
	socketPath string
)

func init() {
	l = log.Logger{}

	flag.BoolVar(&debug, "debug", false, "Enable xDS server debug logging")

	// The port that this xDS server listens on
	flag.StringVar(&socketPath, "socket-path", "/tmp/sds-asm/public/api.sock", "xDS socket path")
}

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	_, _, err := server.Run(socketPath, l)
	if err != nil {
		os.Exit(1)
	}
}

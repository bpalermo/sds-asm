package main

import (
	"flag"
	"github.com/bpalermo/sds-asm/internal/helper"
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

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	flag.Parse()

	srv := setup(debug)

	err := srv.Run(socketPath)
	if err != nil {
		os.Exit(1)
	}
}

func setup(isDebug bool) *server.SdsServer {
	if isDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		l.Debugf("debug log enabled")
	}

	srv, err := server.NewServer(
		helper.GetEnv("AWS_REGION", "us-east-1"),
		helper.GetEnv("AWS_ENDPOINT", ""), l)
	if err != nil {
		os.Exit(1)
	}

	return srv
}

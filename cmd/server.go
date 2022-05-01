package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"

	"network-audio/pkg/logx"
	"network-audio/pkg/server"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(
		&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		},
	)
	logger.SetOutput(os.Stderr)

	addr := fmt.Sprintf("tcp://:%d", 3000)
	slog := logx.Scope(logger, "server")

	svr := server.New(slog, addr)

	go func(svr *server.Server) {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		signal.Ignore(syscall.SIGPIPE)

		<-signalChan
		svr.Close()
	}(svr)

	err := gnet.Run(
		svr, addr,
		gnet.WithMulticore(true),
		gnet.WithLogger(logx.Component(slog, "gnet")),
		gnet.WithTicker(true),
	)

	if err != nil {
		log.Fatal(err)
	}
}

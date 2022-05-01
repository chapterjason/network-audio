package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"network-audio/pkg/client"
	"network-audio/pkg/client/player"
	"network-audio/pkg/logx"
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

	args := os.Args[1:]

	if len(args) < 1 {
		logger.Fatal("Usage: client <host>")
	}

	host := args[0]

	clog := logx.Scope(logger, "client")

	cl := player.NewClock(time.Duration(5) * time.Millisecond)

	p := player.New(
		logx.Component(logger, "player"),
		cl,
	)

	c := client.New(
		clog, cl, p, host,
		client.WithReconnectInterval(time.Second*1),
		client.WithReconnectMaxTimes(30),
	)

	go func(c *client.Client) {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		signal.Ignore(syscall.SIGPIPE)

		<-signalChan
		c.Shutdown()
	}(c)

	c.Run()
}

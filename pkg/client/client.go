package client

import (
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"network-audio/pkg/client/player"
	"network-audio/pkg/logx"
	"network-audio/pkg/messages"
	"network-audio/pkg/timex"
)

type ClientOption func(*Client)

func WithLogger(logger logrus.FieldLogger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithReconnectInterval(interval time.Duration) ClientOption {
	return func(c *Client) {
		c.reconnectInterval = interval
	}
}

func WithReconnectMaxTimes(maxTimes int) ClientOption {
	return func(c *Client) {
		c.reconnectMaxTimes = maxTimes
	}
}

type Client struct {
	gnet.BuiltinEventEngine
	engine     gnet.Engine
	connection gnet.Conn
	logger     logrus.FieldLogger

	address string

	reconnectInterval time.Duration
	reconnectMaxTimes int
	reconnectTimes    int

	player *player.Player

	shutdownChan chan bool
	shutdown     bool

	closeChan chan bool
	closed    bool

	clock *player.Clock
}

func New(logger logrus.FieldLogger, clock *player.Clock, player *player.Player, address string, opts ...ClientOption) *Client {
	c := &Client{
		logger:            logger,
		address:           address,
		reconnectTimes:    0,
		reconnectMaxTimes: 3,
		reconnectInterval: time.Second * 5,
		shutdownChan:      make(chan bool),
		closeChan:         make(chan bool),
		closed:            true,
		shutdown:          false,
		player:            player,
		clock:             clock,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Connect() error {
	c.reconnectTimes++

	glog := logx.Component(c.logger, "gnet")

	gc, err := gnet.NewClient(
		c,
		gnet.WithMulticore(true),
		gnet.WithLogger(glog),
		gnet.WithTicker(true),
	)

	if err != nil {
		return err
	}

	_, err = gc.Dial("tcp", c.address)

	if err != nil {
		return err
	}

	err = gc.Start()

	if err != nil {
		return err
	}

	c.closed = false
	c.shutdown = false
	c.reconnectTimes = 0

	return nil
}

func (c *Client) Run() {
	for {
		if c.shutdown {
			break
		} else if c.closed {
			if c.reconnectTimes >= c.reconnectMaxTimes {
				c.logger.Errorf("try to connect %d times, but failed", c.reconnectMaxTimes)
				break
			}

			err := c.Connect()

			if err != nil {
				c.logger.Errorf("failed to connect, retry in %s: %v", c.reconnectInterval.String(), err)

				time.Sleep(c.reconnectInterval)
				continue
			}
		}

		time.Sleep(time.Second)
	}
}

func (c *Client) OnBoot(engine gnet.Engine) gnet.Action {
	c.engine = engine

	c.logger.Infof("client connected to %s\n", c.address)

	return gnet.None
}

func (c *Client) OnTraffic(con gnet.Conn) gnet.Action {
	msg, err := messages.FromConnection(con)

	if err != nil {
		c.logger.Error(err)

		return gnet.Close
	}

	switch m := msg.(type) {
	case *messages.Audio:
		samples := len(m.Left)
		sampleDuration := c.player.SampleDuration(samples)
		sent := timex.ToTime(m.Time)
		playbackTime := sent.Add(sampleDuration)
		received := c.clock.Now()

		diff := received.Sub(playbackTime)

		if diff > time.Second {
			c.logger.Warnf("dropping audio late by %s", diff.String())
			return gnet.None
		}

		go c.player.Enqueue(m)
	case *messages.Latency:
		c.player.UpdateLatency(m)
	default:
		c.logger.Errorf("unknown message type: %T\n", m)
		return gnet.None
	}

	return gnet.None
}

func (c *Client) OnOpen(con gnet.Conn) ([]byte, gnet.Action) {
	c.connection = con

	c.logger.Infof("connection opened: %s\n", con.RemoteAddr())

	go c.player.Play()

	return nil, gnet.None
}

func (c *Client) OnClose(con gnet.Conn, err error) (action gnet.Action) {
	c.logger.Infof("connection closed: %s\n", con.RemoteAddr())

	c.closeChan <- true

	return gnet.None
}

func (c *Client) OnShutdown(engine gnet.Engine) {
	c.player.Close()

	c.logger.Info("client disconnect")
}

func (c *Client) OnTick() (delay time.Duration, action gnet.Action) {
	select {
	case _ = <-c.shutdownChan:
		c.shutdown = true
		return 0, gnet.Shutdown
	case _ = <-c.closeChan:
		c.closed = true
		return 0, gnet.Close
	default:
		if c.shutdown {
			return 0, gnet.Shutdown
		} else if c.closed {
			return 0, gnet.Close
		}

		tmsg := &messages.Time{
			Time: timex.ToTimestamp(time.Now()),
		}

		err := c.Send(tmsg)

		if err != nil {
			c.logger.Errorf("failed to send time message: %v", err)
		}

		latency := c.clock.GetLatency()

		if latency > time.Second {
			c.logger.Warnf("latency is high %s", latency.String())
		}

		return time.Millisecond * 100, gnet.None
	}
}

func (c *Client) Send(msg proto.Message) error {
	packet := messages.ToPacket(msg)
	bytes, err := packet.Bytes()

	if err != nil {
		return err
	}

	err = c.connection.AsyncWrite(bytes, nil)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Shutdown() {
	c.shutdownChan <- true
}

func (c *Client) Close() {
	c.closeChan <- true
}

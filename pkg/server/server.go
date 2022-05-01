package server

import (
	"sync"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"network-audio/pkg/logx"
	"network-audio/pkg/messages"
	"network-audio/pkg/server/player"
	"network-audio/pkg/timex"
)

type Server struct {
	gnet.BuiltinEventEngine

	engine gnet.Engine
	log    *logrus.Entry

	address string
	logger  logrus.FieldLogger

	player *player.Player

	clients  *sync.Map
	stopChan chan bool

	counter uint64
}

func New(logger logrus.FieldLogger, address string) *Server {
	s := &Server{
		logger:   logger,
		address:  address,
		clients:  &sync.Map{},
		stopChan: make(chan bool),
	}

	s.player = player.New(
		s,
		logx.Component(logger, "player"),
	)

	return s
}

func (s *Server) OnBoot(engine gnet.Engine) gnet.Action {
	s.engine = engine

	s.logger.Infof("server is listening on %s\n", s.address)

	// loop the file player
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			default:
				wg := &sync.WaitGroup{}

				wg.Add(1)

				go func() {
					s.logger.Info("start to play file")
					defer s.player.Stop()
					err := s.player.PlayFile("./test/audio.mp3")

					if err != nil {
						s.logger.Errorf("play file error: %s\n", err)
					}

					wg.Done()
				}()

				wg.Wait()

				s.logger.Info("file play done")

				time.Sleep(time.Second)
			}
		}
	}()

	return gnet.None
}

func (s *Server) OnShutdown(engine gnet.Engine) {
	s.player.Stop()

	s.logger.Info("server is shutdown")
}

func (s *Server) OnTraffic(c gnet.Conn) gnet.Action {
	msg, err := messages.FromConnection(c)

	if err != nil {
		s.logger.Errorf("read message error: %s\n", err)

		return gnet.None
	}

	switch m := msg.(type) {
	case *messages.Time:
		sent := timex.ToTime(m.Time)
		received := time.Now()
		latency := received.Sub(sent)

		msg := &messages.Latency{
			Latency: latency.Nanoseconds(),
			Time:    timex.ToTimestamp(time.Now()),
		}

		err := s.SendTo(c, msg)

		if err != nil {
			return gnet.Close
		}
	default:
		s.logger.Errorf("unknown message type: %T\n", m)
		return gnet.None
	}

	return gnet.None
}

func (s *Server) OnOpen(connection gnet.Conn) ([]byte, gnet.Action) {
	remoteAddr := connection.RemoteAddr().String()

	s.logger.Infof("connection opened: %s\n", remoteAddr)
	s.clients.Store(remoteAddr, connection)

	return nil, gnet.None
}

func (s *Server) OnClose(connection gnet.Conn, err error) (action gnet.Action) {
	remoteAddr := connection.RemoteAddr().String()

	s.logger.Infof("connection closed: %s\n", remoteAddr)
	s.clients.Delete(remoteAddr)

	return gnet.None
}

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	select {
	case _ = <-s.stopChan:
		return 0, gnet.Shutdown
	default:
		return 0, gnet.None
	}
}

func (s *Server) Close() {
	s.stopChan <- true
}

func (s *Server) Send(msg proto.Message) error {
	packet := messages.ToPacket(msg)
	bytes, err := packet.Bytes()

	if err != nil {
		return err
	}

	s.Broadcast(bytes)

	return nil
}

func (s *Server) SendTo(connection gnet.Conn, msg proto.Message) error {
	packet := messages.ToPacket(msg)
	bytes, err := packet.Bytes()

	if err != nil {
		s.logger.Errorf("packet to bytes error: %s\n", err)
		return err
	}

	_, err = connection.Write(bytes)

	if err != nil {
		s.logger.Errorf("error writing to connection: %s\n", err)
		return err
	}

	return nil
}

func (s *Server) Broadcast(bytes []byte) {
	s.clients.Range(
		func(key, value interface{}) bool {

			go func(connection gnet.Conn) {
				_, err := connection.Write(bytes)

				if err != nil {
					s.logger.Errorf("error writing to connection: %s\n", err)
				}
			}(value.(gnet.Conn))

			return true
		},
	)
}

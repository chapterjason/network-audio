package player

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/sirupsen/logrus"

	"network-audio/pkg/messages"
	"network-audio/pkg/timex"
)

type Player struct {
	logger logrus.FieldLogger
	target Target

	format           beep.Format
	streamBufferSize int
	resampleQuality  int

	stopChan chan bool
}

type Option func(*Player)

func WithFormat(format beep.Format) Option {
	return func(p *Player) {
		p.format = format
	}
}

func WithResampleQuality(resampleQuality int) Option {
	return func(p *Player) {
		p.resampleQuality = resampleQuality
	}
}

func New(target Target, logger logrus.FieldLogger, options ...Option) *Player {
	p := &Player{
		logger:           logger,
		format:           beep.Format{SampleRate: 44100, Precision: 2, NumChannels: 2},
		resampleQuality:  3,
		streamBufferSize: 512,
		target:           target,
		stopChan:         make(chan bool),
	}

	for _, option := range options {
		option(p)
	}

	p.logger.Infof("Player: streamBufferSize: %d", p.streamBufferSize)

	return p
}

// PlayFile is a blocking function that plays a file.
func (p *Player) PlayFile(filePath string) error {
	p.stopChan = make(chan bool)

	stream, err := p.getFileStream(filePath)

	if err != nil {
		return err
	}

	return p.playStream(stream)
}

func (p *Player) getFileStream(filePath string) (beep.Streamer, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	stream, format, err := mp3.Decode(file)

	if err != nil {
		return nil, err
	}

	if p.format.SampleRate != format.SampleRate {
		return beep.Resample(p.resampleQuality, format.SampleRate, p.format.SampleRate, stream), nil
	}

	return stream, nil
}

func (p *Player) Stop() {
	p.stopChan <- true
}

func (p *Player) playStream(stream beep.Streamer) error {
	buffer := make([][2]float64, p.streamBufferSize)
	ok := true
	samplesAmount := p.streamBufferSize

LOOP:
	for {
		select {
		case _ = <-p.stopChan:
			ok = false
		default:
			iterationStart := time.Now()

			if ok != true {
				break LOOP
			}

			samplesAmount, ok = stream.Stream(buffer)

			samplesLeft := make([]float64, samplesAmount)
			samplesRight := make([]float64, samplesAmount)

			for i := 0; i < samplesAmount; i++ {
				samplesLeft[i] = buffer[i][0]
				samplesRight[i] = buffer[i][1]
			}

			playbackInterval := p.format.SampleRate.D(samplesAmount)

			err := p.target.Send(
				&messages.Audio{
					Left:  samplesLeft,
					Right: samplesRight,
					Time:  timex.ToTimestamp(time.Now().Add(time.Nanosecond * 100)),
				},
			)

			if err != nil {
				return err
			}

			time.Sleep(playbackInterval - time.Since(iterationStart) - time.Nanosecond*100)
		}
	}

	return nil
}

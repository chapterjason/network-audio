package player

import (
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/sirupsen/logrus"

	"network-audio/pkg/audio"
	"network-audio/pkg/circularbuffer"
	"network-audio/pkg/messages"
	"network-audio/pkg/timex"
)

type Player struct {
	logger logrus.FieldLogger

	streamBuffer *circularbuffer.Queue

	// Will be used to fill the player buffer if the available data is not enough
	fillStreamer beep.Streamer

	format beep.Format

	// The buffer used by the beep.Speaker
	bufferSize int

	clock *Clock

	// This threshold defines the maximum distance between the next packet and now
	delayThreshold time.Duration
}

type Option func(*Player)

func WithFormat(format beep.Format) Option {
	return func(p *Player) {
		p.format = format
	}
}

func WithDelayThreshold(delayThreshold time.Duration) Option {
	return func(p *Player) {
		p.delayThreshold = delayThreshold
	}
}

func WithBufferSize(bufferSize int) Option {
	return func(p *Player) {
		p.bufferSize = bufferSize
	}
}

func WithFillStreamer(fillStreamer beep.Streamer) Option {
	return func(p *Player) {
		p.fillStreamer = fillStreamer
	}
}

func New(logger logrus.FieldLogger, clock *Clock, opts ...Option) *Player {
	p := &Player{
		logger:         logger,
		format:         beep.Format{SampleRate: 44100, Precision: 2, NumChannels: 2},
		delayThreshold: time.Millisecond * 10,
		bufferSize:     256,

		// endless silence streamer
		fillStreamer: beep.Silence(-1),

		clock: clock,
	}

	for _, opt := range opts {
		opt(p)
	}

	// create a buffer with a size of 10 seconds of audio based on the format
	streamBufferSize := p.format.SampleRate.N(time.Millisecond * 500)
	p.streamBuffer = circularbuffer.New(streamBufferSize)

	p.logger.Infof("Client: streamBufferSize: %d", streamBufferSize)
	p.logger.Infof("Client: playerBufferSize: %d", p.bufferSize)

	return p
}

func (p *Player) SampleDuration(n int) time.Duration {
	return p.format.SampleRate.D(n)
}

func (p *Player) UpdateLatency(m *messages.Latency) time.Duration {
	return p.clock.UpdateLatency(m)
}

func (p *Player) Enqueue(am *messages.Audio) {
	for index, leftData := range am.Left {
		data := [2]float64{
			leftData,
			am.Right[index],
		}

		offset := p.SampleDuration(index)
		s := audio.New(data, timex.ToTime(am.Time).Add(offset))

		if time.Since(s.Time) > time.Millisecond*100 {
			break
		}

		p.streamBuffer.Enqueue(s)
	}
}

func (p *Player) Play() {
	_ = speaker.Init(p.format.SampleRate, p.bufferSize)

	speaker.Play(
		beep.Seq(
			beep.Callback(
				func() {
					p.logger.Info("Start playing...")
				},
			),
			p,
			beep.Callback(
				func() {
					p.logger.Info("Shutdown playing...")
				},
			),
		),
	)
}

func (p *Player) Close() {
	p.streamBuffer.Clear()
	speaker.Clear()
}

func (p *Player) Stream(samples [][2]float64) (n int, ok bool) {
	requiredSamples := len(samples)
	fillSamples := 0
	droppedSamples := 0

	now := p.clock.Now()
	as := p.streamBuffer.Peek().(*audio.Sample)

	// drop samples until the first sample is in the delay threshold
	for now.Sub(as.Time) >= p.delayThreshold {
		droppedSamples++
		_ = p.streamBuffer.Dequeue().(*audio.Sample)
		now = p.clock.Now()
		as = p.streamBuffer.Peek().(*audio.Sample)
	}

	// we are behind the delay threshold, so we need to fill the buffer with samples from the fillStreamer
	diff := as.Time.Sub(now)

	if diff > p.delayThreshold {
		fillSamples = p.format.SampleRate.N(diff)

		// take the lower one
		if requiredSamples < fillSamples {
			fillSamples = requiredSamples
		}

		p.fillStreamer.Stream(samples[:fillSamples])
	}

	for i := fillSamples; i < requiredSamples; i++ {
		as := p.streamBuffer.Dequeue()

		if as == nil {
			p.logger.Warnf("buffer underflow, dropping samples")
			break
		}

		samples[i] = as.(*audio.Sample).Data
	}

	if fillSamples > 0 {
		p.logger.Warnf("filled %d samples", droppedSamples)
	}

	if droppedSamples > 0 {
		p.logger.Warnf("dropped %d samples", droppedSamples)
	}

	return len(samples), len(samples) > 0
}

// Err The player should never malfunction, so Err always returns nil
func (p *Player) Err() error {
	return nil
}

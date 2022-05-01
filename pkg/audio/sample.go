package audio

import (
	"time"
)

// Sample represents a single sample of audio data.
type Sample struct {
	// The samples for each channel.
	Data [2]float64
	Time time.Time
}

func New(data [2]float64, time time.Time) *Sample {
	return &Sample{data, time}
}

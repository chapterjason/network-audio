package messages

import (
	"google.golang.org/protobuf/proto"
)

// Codes for the different types of messages.
const (
	AudioType   = 0x10
	TimeType    = 0x20
	LatencyType = 0x30
)

func ToPacket(message proto.Message) *Packet {
	packet := &Packet{}

	switch message.(type) {
	case *Audio:
		packet.mtype = AudioType
	case *Time:
		packet.mtype = TimeType
	case *Latency:
		packet.mtype = LatencyType
	default:
		panic("unsupported message type")
	}

	packet.message = message

	return packet
}

// ToBytes marshals a messageChan to a buffer. The first byte of the buffer is the messageChan type.
func ToBytes(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)

	if err != nil {
		return nil, err
	}

	return data, nil
}

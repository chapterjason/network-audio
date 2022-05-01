package messages

import (
	"encoding/binary"
	"fmt"

	"github.com/panjf2000/gnet/v2"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// Packet is a wrapper around a byte array.
//
// Format:
// -----------------------------------
// | Message Type | Length  | Message |
// ------------------------------------
// |   8 bytes    | 8 bytes | N bytes |
// ------------------------------------
// |	int       | uint64  |  []byte |
// ------------------------------------
//
type Packet struct {
	mtype   int
	message proto.Message
}

func FromConnection(connection gnet.Conn) (proto.Message, error) {
	headerBytes, err := connection.Peek(16)

	if err != nil {
		return nil, fmt.Errorf("error reading type byte: %v", err)
	}

	messageType := int(binary.BigEndian.Uint64(headerBytes[0:8]))
	messageLength := int(binary.BigEndian.Uint64(headerBytes[8:16]))

	var message proto.Message

	switch messageType {
	case AudioType:
		message = &Audio{}
	case TimeType:
		message = &Time{}
	case LatencyType:
		message = &Latency{}
	default:
		return nil, fmt.Errorf("unsupported message type: %v", messageType)
	}

	_, err = connection.Discard(16)

	if err != nil {
		return nil, errors.Wrap(err, "error discarding header bytes")
	}

	peekedBytes, err := connection.Peek(messageLength)

	if err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}

	if len(peekedBytes) != messageLength {
		return nil, fmt.Errorf("error reading message: expected %v bytes, got %v", messageLength, len(peekedBytes))
	}

	messageBytes := make([]byte, messageLength)

	bytesRead, err := connection.Read(messageBytes)

	if err != nil {
		return nil, fmt.Errorf("error reading message: %v", err)
	}

	if bytesRead != messageLength {
		return nil, fmt.Errorf("error reading message: expected %v bytes, got %v", messageLength, bytesRead)
	}

	err = proto.Unmarshal(messageBytes, message)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling message: %v", err)
	}

	return message, nil
}

func (p *Packet) Message() proto.Message {
	return p.message
}

func (p *Packet) HeaderBytes() ([]byte, error) {
	mts := p.MessageTypeBytes()
	mlb, err := p.MessageLengthBytes()

	if err != nil {
		return nil, err
	}

	b := make([]byte, 16)
	copy(b[0:8], mts)
	copy(b[8:16], mlb)

	return b, nil
}

func (p *Packet) MessageTypeBytes() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(p.mtype))

	return b
}

func (p *Packet) MessageBytes() ([]byte, error) {
	mb, err := ToBytes(p.message)

	if err != nil {
		return nil, err
	}

	return mb, nil
}

func (p *Packet) Bytes() ([]byte, error) {
	hb, err := p.HeaderBytes()

	if err != nil {
		return nil, err
	}

	mb, err := p.MessageBytes()

	if err != nil {
		return nil, err
	}

	return append(hb, mb...), nil
}

func (p *Packet) MessageLengthBytes() ([]byte, error) {
	mb, err := p.MessageBytes()

	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(len(mb)))

	return b, nil
}

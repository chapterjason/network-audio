package player

import (
	"google.golang.org/protobuf/proto"
)

type Target interface {
	Send(msg proto.Message) error
}

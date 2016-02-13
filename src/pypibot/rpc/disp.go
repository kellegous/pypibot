package rpc

import (
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
)

const (
	msgPingMsg uint32 = iota
)

func cmdPing(c net.Conn, b []byte) error {
	var m PingReq
	if err := proto.Unmarshal(b, &m); err != nil {
		return err
	}
	return writeMsg(c, msgPingMsg, &PingRes{
		Id: m.Id,
	})
}

func dispatch(c net.Conn, t uint32, b []byte) error {
	switch t {
	case msgPingMsg:
		return cmdPing(c, b)
	default:
		return fmt.Errorf("invalid message: %d", t)
	}
}

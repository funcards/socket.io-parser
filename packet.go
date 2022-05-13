package sio_parser

import "strconv"

type (
	// PacketType indicates type of socket.io Packet
	PacketType byte

	Packet struct {
		Type        PacketType
		Nsp         string
		Data        any
		ID          *uint64
		Attachments int
	}
)

const (
	Connect PacketType = iota
	Disconnect
	Event
	Ack
	ConnectError
	BinaryEvent
	BinaryAck
)

const (
	StrConnect      = "CONNECT"
	StrDisconnect   = "DISCONNECT"
	StrEvent        = "EVENT"
	StrAck          = "ACK"
	StrConnectError = "CONNECT_ERROR"
	StrBinaryEvent  = "BINARY_EVENT"
	StrBinaryAck    = "BINARY_ACK"
)

var (
	PacketTypeToString = map[PacketType]string{
		Connect:      StrConnect,
		Disconnect:   StrDisconnect,
		Event:        StrEvent,
		Ack:          StrAck,
		ConnectError: StrConnectError,
		BinaryEvent:  StrBinaryEvent,
		BinaryAck:    StrBinaryAck,
	}
	StringToPacketType = map[string]PacketType{
		StrConnect:      Connect,
		StrDisconnect:   Disconnect,
		StrEvent:        Event,
		StrAck:          Ack,
		StrConnectError: ConnectError,
		StrBinaryEvent:  BinaryEvent,
		StrBinaryAck:    BinaryAck,
	}
)

// String returns string representation of a PacketType
func (p PacketType) String() string {
	return PacketTypeToString[p]
}

func (p PacketType) Int() int {
	return int(p)
}

func (p PacketType) Encode() string {
	return strconv.Itoa(p.Int())
}

func (p PacketType) Bytes() []byte {
	return []byte(p.Encode())
}

func (p Packet) Encode() []any {
	return Encode(p)
}

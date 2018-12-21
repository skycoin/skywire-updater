package transport

import (
	"encoding/binary"
	"math"
)

// RouteID represents ID of a Route in a Packet.
type RouteID uint32

// InitRouteID represents RouteID of the packet that is used to
// establish Routing Rules, Transport Registration, etc.
const InitRouteID = RouteID(0)

// Packet defines generic packet recognized by all skywire nodes.
type Packet []byte

// MakePacket constructs a new Packet. If payload size is more than
// uint16, PutUvarint will panic.
func MakePacket(id RouteID, payload []byte) Packet {
	if len(payload) > math.MaxUint16 {
		panic("packet size exceeded")
	}

	packet := make([]byte, 6)
	binary.LittleEndian.PutUint32(packet, uint32(id))
	binary.LittleEndian.PutUint16(packet[4:], uint16(len(payload)))
	return Packet(append(packet, payload...))
}

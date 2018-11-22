package discovery

// PacketType is the type of the packets are sent within transports (between Skywire Nodes).
type PacketType int

const (
	Ping           PacketType = 0x0  // Ping sent between a Transport to check if connection is still open, and to determine latency of the Transport.
	InitiateRoute             = 0x01 // InitiateRoute is the first packet sent via a route to have it initiated.
	RouteInitiated            = 0x02 // RouteInitiated confirms that a route is set up and functional.
	DestroyRoute              = 0x03 // DestroyRoute initiate the destruction of a route.
	RouteDestroyed            = 0x04 // RouteDestroyed confirm the success of a route's destruction.
	OpenStream                = 0x05 // OpenStream opens a stream within a route.
	StreamOpened              = 0x06 // StreamOpened confirms that a stream is successfully opened.
	CloseStream               = 0x07 // CloseStream closes a stream.
	StreamClosed              = 0x06 // StreamClosed informs that a stream has successfully closed.
	Forward                   = 0x07 // Forward forwards data via a specified stream.
)

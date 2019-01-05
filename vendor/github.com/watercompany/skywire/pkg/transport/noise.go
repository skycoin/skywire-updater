package transport

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/flynn/noise"
	"github.com/skycoin/skycoin/src/cipher"

	"github.com/watercompany/skywire/internal/dh"
)

func init() {
	cipher.DebugLevel2 = false // DebugLevel2 causes ECDH to be really slow
}

const (
	// packetsTillRekey is the number of packages after which we want to rekey for the noise protocol
	packetsTillRekey = 10

	// DefaultHandshakeTimeout since for now it not specified by caller
	DefaultHandshakeTimeout = time.Second * 3

	// XKFrameASize is the handshake's frame A size.
	XKFrameASize = 49

	// XKFrameBSize is the handshake's frame B size.
	XKFrameBSize = 49

	// XKFrameCSize is the handshake's frame C size.
	XKFrameCSize = 65
)

// Noise transport errors
var (
	ErrHandshakeFailed               = errors.New("handshake failed")
	ErrTransportCommunicationTimeout = errors.New("transport communication operation timed out")
)

type noiseTransportConfig struct {
	StaticPublic     cipher.PubKey // Local instance static public key.
	StaticSecret     cipher.SecKey // Local instance static secret key.
	StaticRemote     cipher.PubKey // Remote instance static public key.
	Initiator        bool          // Whether the local instance initiates the connection.
	HandshakeTimeout time.Duration
	PingInterval     time.Duration
}

// NoiseTransport implements Transport.
type NoiseTransport struct {
	originalTransport Transport
	deadLine          time.Time
	noise             *noiseInstance
	readChan          chan []byte
	buf               *bytes.Buffer
	sync.Mutex
}

// NewNoiseTransport adds a layer of noiseInstance encryption on top of an
// unencrypted Transport implementation.
// If the noiseInstance handshake fails, an error is to be returned.
func NewNoiseTransport(secretKey cipher.SecKey, transport Transport, initiator bool) (*NoiseTransport, error) {
	// Definition ...
	c := &noiseTransportConfig{
		StaticSecret: secretKey,
		StaticPublic: transport.Local(),
		StaticRemote: transport.Remote(),
		Initiator:    initiator,
	}
	noiseInstance, err := newNoiseWithXKAndSecp256k1(c)
	if err != nil {
		return nil, err
	}

	handshake := noiseInstance.MakeHandshake()
	err = handshake.Do(transport, DefaultHandshakeTimeout)
	if err != nil {
		return nil, err
	}

	nt := &NoiseTransport{
		originalTransport: transport,
		deadLine:          time.Now().Add(time.Hour),
		noise:             noiseInstance,
		buf:               &bytes.Buffer{},
		readChan:          make(chan []byte),
	}
	return nt, nil
}

// Read implements io.Reader for noise encryption transport
func (t *NoiseTransport) Read(p []byte) (n int, err error) {
	if t.buf.Len() != 0 {
		return t.buf.Read(p)
	}

	packetSize := make([]byte, 2)
	n, err = t.originalTransport.Read(packetSize)
	if err != nil {
		return n, err
	}
	size := binary.BigEndian.Uint16(packetSize)

	readBuffer := make([]byte, size)
	n, err = t.originalTransport.Read(readBuffer)
	if err != nil {
		return n, err
	}

	readed, err := t.noise.Decrypt(readBuffer, nil)
	if err != nil {
		return 0, err
	}
	if len(readed) > len(p) {
		if _, err := t.buf.Write(readed[len(p):]); err != nil {
			return 0, io.ErrShortBuffer
		}

		return copy(p, readed[:len(p)]), nil
	}

	return copy(p, readed), nil
}

// Write implements io.Writer for noise encryption transport
// The two first bytes of the message is a BigEndian integer representing the size of the rest of
// the message, which is chacha-encoded p
func (t *NoiseTransport) Write(p []byte) (n int, err error) {
	encoded := t.noise.Encrypt(p, nil)
	size := len(encoded)
	packet := make([]byte, size+2)
	binary.BigEndian.PutUint16(packet, uint16(size))
	copy(packet[2:], encoded)
	n, err = t.originalTransport.Write(packet)

	return n - (len(encoded) - len(p)) - 2, err
}

// Close implements io.Closer for noise encryption transport
func (t *NoiseTransport) Close() error {
	return t.originalTransport.Close()
}

// Local returns the local transport edge's public key.
func (t *NoiseTransport) Local() cipher.PubKey {
	return t.originalTransport.Local()
}

// Remote returns the remote transport edge's public key.
func (t *NoiseTransport) Remote() cipher.PubKey {
	return t.originalTransport.Remote()
}

// SetDeadline functions the same as that from net.Conn
// With a Transport, we don't have a distinction between write and read timeouts.
func (t *NoiseTransport) SetDeadline(date time.Time) error {
	// if I call cancel here the whole context will be canceled after setting its new deadline
	return t.originalTransport.SetDeadline(date)
}

// Type returns the string representation of the transport type.
func (t *NoiseTransport) Type() string {
	return "noise"
}

// Handshake represents a set of actions that an instance performs to complete a handshake.
type Handshake func(rw io.ReadWriter) error

// Do performs a handshake with a given timeout.
// Non-nil error is returned on failure.
func (handshake Handshake) Do(rw io.ReadWriter, timeout time.Duration) (err error) {
	var (
		done = make(chan error, 1) // handshake response channel
	)
	go func() {
		done <- handshake(rw)
	}()
	select {
	case <-time.After(timeout):
		err = ErrHandshakeFailed
	case err = <-done:
	}
	return err
}

// noiseInstance handles the handshake and the frame's cryptography.
type noiseInstance struct {
	config               *noiseTransportConfig
	handshake            *noise.HandshakeState
	encrypt              *noise.CipherState
	decrypt              *noise.CipherState
	encryptPacketCounter uint16
	decryptPacketCounter uint16
	encryptLock          sync.RWMutex
	decryptLock          sync.RWMutex
}

// newNoiseWithXKAndSecp256k1 creates a new noiseInstance instance with:
//	- XK pattern for handshake.
//	- Secp256k1 for the curve.
func newNoiseWithXKAndSecp256k1(nc *noiseTransportConfig) (*noiseInstance, error) {
	var (
		noisePattern     = noise.HandshakeXK
		noiseCipherSuite = noise.NewCipherSuite(dh.Secp256k1{}, noise.CipherChaChaPoly, noise.HashSHA256)
	)
	config := noise.Config{
		CipherSuite: noiseCipherSuite,
		Random:      rand.Reader,
		Pattern:     noisePattern,
		Initiator:   nc.Initiator,
		StaticKeypair: noise.DHKey{
			Public:  nc.StaticPublic[:],
			Private: nc.StaticSecret[:],
		},
	}
	if nc.Initiator {
		config.PeerStatic = nc.StaticRemote[:]
	}
	hs, err := noise.NewHandshakeState(config)
	if err != nil {
		return nil, err
	}
	return &noiseInstance{
		config:      nc,
		handshake:   hs,
		encryptLock: sync.RWMutex{},
		decryptLock: sync.RWMutex{},
	}, nil
}

// MakeHandshake creates a noise handshake.
func (ns *noiseInstance) MakeHandshake() Handshake {
	var handshake Handshake
	if ns.config.Initiator {
		// Handshake actions of the initiating instance.
		handshake = func(rw io.ReadWriter) error {
			// Send A.
			if _, err := ns.writeHandshakeFrame(rw); err != nil {
				return err
			}
			// Receive B.
			if _, err := ns.readHandshakeFrame(rw, XKFrameBSize); err != nil {
				return err
			}
			// Send C.
			if _, err := ns.writeHandshakeFrame(rw); err != nil {
				return err
			}
			// Conclude handshake.
			return ns.finishHandshake()
		}
	} else {
		// Handshake actions of the responding instance.
		handshake = func(rw io.ReadWriter) error {
			// Receive A.
			if _, err := ns.readHandshakeFrame(rw, XKFrameASize); err != nil {
				return err
			}
			// Send B.
			if _, err := ns.writeHandshakeFrame(rw); err != nil {
				return err
			}
			// Receive C.
			if _, err := ns.readHandshakeFrame(rw, XKFrameCSize); err != nil {
				return err
			}
			// Conclude handshake.
			return ns.finishHandshake()
		}
	}
	return handshake
}

// Encrypt encrypts a frame after the handshake is successfully completed.
func (ns *noiseInstance) Encrypt(in, ad []byte) []byte {
	ns.encryptCountOrRekey()

	ns.encryptLock.RLock()
	v := ns.encrypt.Encrypt(nil, ad, in)
	ns.encryptLock.RUnlock()

	return v
}

// Decrypt decrypts a frame after the handshake is successfully completed.
func (ns *noiseInstance) Decrypt(in, ad []byte) ([]byte, error) {
	ns.decryptCountOrRekey()

	ns.decryptLock.RLock()
	b, err := ns.decrypt.Decrypt(nil, ad, in)
	ns.decryptLock.RUnlock()

	return b, err
}

func (ns *noiseInstance) encryptCountOrRekey() {
	ns.encryptLock.Lock()
	defer ns.encryptLock.Unlock()

	ns.encryptPacketCounter++
	if ns.encryptPacketCounter >= packetsTillRekey {
		ns.encryptPacketCounter = 0
		ns.encrypt.Rekey()
	}
}

func (ns *noiseInstance) decryptCountOrRekey() {
	ns.decryptLock.Lock()
	defer ns.decryptLock.Unlock()

	ns.decryptPacketCounter++
	if ns.decryptPacketCounter >= packetsTillRekey {
		ns.decryptPacketCounter = 0
		ns.decrypt.Rekey()
	}
}

// nolint
func (ns *noiseInstance) readHandshakeFrame(reader io.Reader, packetSize int) (int, error) {
	packet := make([]byte, packetSize)
	n, err := io.ReadFull(reader, packet)
	if err != nil {
		return n, err
	}
	if ns.config.Initiator {
		_, ns.encrypt, ns.decrypt, err = ns.handshake.ReadMessage(nil, packet)
	} else {
		_, ns.decrypt, ns.encrypt, err = ns.handshake.ReadMessage(nil, packet)
	}
	return n, err
}

// nolint
func (ns *noiseInstance) writeHandshakeFrame(writer io.Writer) (int, error) {
	var (
		packet []byte
		err    error
	)
	if ns.config.Initiator {
		packet, ns.encrypt, ns.decrypt, err = ns.handshake.WriteMessage(nil, nil)
	} else {
		packet, ns.decrypt, ns.encrypt, err = ns.handshake.WriteMessage(nil, nil)
	}
	if err != nil {
		return 0, err
	}
	return writer.Write(packet)
}

func (ns *noiseInstance) finishHandshake() error {
	if !ns.config.Initiator {
		copy(ns.config.StaticRemote[:33], ns.handshake.PeerStatic()[:33])
	}
	ns.handshake = nil
	return nil
}

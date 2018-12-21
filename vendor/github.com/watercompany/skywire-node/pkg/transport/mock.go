package transport

import (
	"context"
	"io"
	"sync"
	"time"

	"errors"

	"github.com/skycoin/skycoin/src/cipher"
)

// errors from mock transport
var (
	ErrPKNotRegistered = errors.New("transport with this public key is not registered")
)

type pipes struct {
	static cipher.PubKey
	reader *io.PipeReader
	notify chan *pipes
}

func newPipes(pk cipher.PubKey, reader *io.PipeReader) *pipes {
	return &pipes{
		static: pk,
		reader: reader,
		notify: make(chan *pipes),
	}
}

type pipeDeliverer struct {
	pipes map[cipher.PubKey]*pipes
	sync.RWMutex
}

// nolint used in tests
func newPipeDeliverer() *pipeDeliverer {
	return &pipeDeliverer{
		pipes: make(map[cipher.PubKey]*pipes),
	}
}

func (p *pipeDeliverer) RegisterAndBlock(pk cipher.PubKey, reader *io.PipeReader) *pipes {
	pipes := newPipes(pk, reader)
	p.Lock()
	p.pipes[pk] = pipes
	p.Unlock()

	return <-pipes.notify
}

func (p *pipeDeliverer) Lookup(pk cipher.PubKey, reader *io.PipeReader) (*io.PipeReader, error) {
	p.RLock()
	pipe, ok := p.pipes[pk]
	p.RUnlock()

	if !ok {
		return nil, ErrPKNotRegistered
	}

	p.Lock()
	delete(p.pipes, pk)
	p.Unlock()

	pipe.notify <- newPipes(pk, reader)

	return pipe.reader, nil
}

// pipeFactory spans transports that communicate trhoug io.PipeReader and io.PipeWriter
type pipeFactory struct {
	pipeDeliverer *pipeDeliverer
	staticSecret  cipher.SecKey
	staticPublic  cipher.PubKey
	writer        *io.PipeWriter
	reader        *io.PipeReader
}

// nolint used in tests
func newPipeFactory(sk cipher.SecKey, deliverer *pipeDeliverer) *pipeFactory {
	reader, writer := io.Pipe()

	return &pipeFactory{
		pipeDeliverer: deliverer,
		staticSecret:  sk,
		staticPublic:  cipher.PubKeyFromSecKey(sk),
		reader:        reader,
		writer:        writer,
	}
}

// Accept accepts a remotely-initiated Transport.
func (p *pipeFactory) Accept(ctx context.Context) (Transport, error) {
	remotePipes := p.pipeDeliverer.RegisterAndBlock(p.staticPublic, p.reader)

	return NewMockTransport(p.staticSecret, remotePipes.static,
		p.writer, remotePipes.reader), nil
}

// Dial initiates a Transport with a remote node.
func (p *pipeFactory) Dial(ctx context.Context, remote cipher.PubKey) (Transport, error) {
	remoteReader, err := p.pipeDeliverer.Lookup(remote, p.reader)
	if err != nil {
		return nil, err
	}

	return NewMockTransport(p.staticSecret, remote, p.writer, remoteReader), nil
}

// Close implements closer
func (p *pipeFactory) Close() error {
	p.reader.Close()
	p.writer.Close()

	return nil
}

// Local returns the local public key.
func (p *pipeFactory) Local() cipher.PubKey {
	return p.staticPublic
}

// Type returns the Transport type.
func (p *pipeFactory) Type() string {
	return "pipe"
}

// MockTransport is a transport that accepts custom writers and readers to use them in Read and Write
// operations
type MockTransport struct {
	writer  io.Writer
	reader  io.Reader
	pk      cipher.PubKey
	sk      cipher.SecKey
	remote  cipher.PubKey
	context context.Context
}

// NewMockTransport creates a transport with the given secret key and remote public key, taking a writer
// and a reader that will be used in the Write and Read operation
func NewMockTransport(sk cipher.SecKey, remote cipher.PubKey, writer io.Writer, reader io.Reader) *MockTransport {
	return &MockTransport{
		writer:  writer,
		reader:  reader,
		pk:      cipher.PubKeyFromSecKey(sk),
		sk:      sk,
		remote:  remote,
		context: context.Background(),
	}
}

// Read implements reader for mock transport
func (m *MockTransport) Read(p []byte) (n int, err error) {
	select {
	case <-m.context.Done():
		return 0, ErrTransportCommunicationTimeout
	default:
		return m.reader.Read(p)
	}
}

// Write implements writer for mock transport
func (m *MockTransport) Write(p []byte) (n int, err error) {
	select {
	case <-m.context.Done():
		return 0, ErrTransportCommunicationTimeout
	default:
		return m.writer.Write(p)
	}
}

// Close implements closer for mock transport
func (m *MockTransport) Close() error {
	return nil
}

// Local returns the local static public key
func (m *MockTransport) Local() cipher.PubKey {
	return m.pk
}

// Remote returns the remote public key fo the mock transport
func (m *MockTransport) Remote() cipher.PubKey {
	return m.remote
}

// SetDeadline sets a deadline for the write/read operations of the mock transport
func (m *MockTransport) SetDeadline(t time.Time) error {
	// nolint
	ctx, cancel := context.WithDeadline(m.context, t)
	m.context = ctx

	go func(cancel context.CancelFunc) {
		time.Sleep(time.Until(t))
		cancel()
	}(cancel)

	return nil
}

// Type returns the type of the mock transport
func (m *MockTransport) Type() string {
	return "mock"
}

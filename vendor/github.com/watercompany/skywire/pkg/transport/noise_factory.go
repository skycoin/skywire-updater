package transport

import (
	"context"
	"errors"

	"github.com/skycoin/skycoin/src/cipher"
)

// FactoryErrors
var (
	ErrIsNotAConnecter = errors.New("underlying interface is not a server connecter")
)

// NoiseFactory spans a noise transport either by accepting a remotely initiated connection or initiating
// one to a remote client
type NoiseFactory struct {
	Factory
	transportConfig *noiseTransportConfig
}

// NewNoiseFactory creates a new noise factory from a secret key and another factory implementation
func NewNoiseFactory(sk cipher.SecKey, factory Factory) *NoiseFactory {
	pk, err := cipher.PubKeyFromSecKey(sk)
	if err != nil {
		panic(err) // TODO(evanlinjin): Is panicing the best solution here?
	}

	return &NoiseFactory{
		transportConfig: &noiseTransportConfig{
			StaticPublic: pk,
			StaticSecret: sk,
		},
		Factory: factory,
	}
}

// Accept accepts a remotely-initiated Transport.
func (nf *NoiseFactory) Accept(ctx context.Context) (Transport, error) {
	transport, err := nf.Factory.Accept(ctx)
	if err != nil {
		return nil, err
	}

	noiseTransport, err := NewNoiseTransport(nf.transportConfig.StaticSecret, transport, false)
	if err != nil {
		return nil, err
	}

	return noiseTransport, nil
}

// Dial initiates a Transport with a remote node.
func (nf *NoiseFactory) Dial(ctx context.Context, remote cipher.PubKey) (Transport, error) {
	transport, err := nf.Factory.Dial(ctx, remote)
	if err != nil {
		return nil, err
	}

	noiseTransport, err := NewNoiseTransport(nf.transportConfig.StaticSecret, transport, true)
	if err != nil {
		return nil, err
	}

	return noiseTransport, nil
}

// Close implements io.Closer.
func (nf *NoiseFactory) Close() error {
	return nf.Factory.Close()
}

// InitialServerConnecter has a method that allows to connect to servers on startup. It is sometimes
// necessary in order for factory to spawn transports.
type InitialServerConnecter interface {
	ConnectToInitialServers(ctx context.Context, serverCount int) error
}

// ConnectToInitialServers implements InitialServerConnecter
func (nf *NoiseFactory) ConnectToInitialServers(ctx context.Context, serverCount int) error {
	if isc, ok := nf.Factory.(InitialServerConnecter); ok {
		return isc.ConnectToInitialServers(ctx, serverCount)
	}

	return ErrIsNotAConnecter
}

// Local returns the local public key.
func (nf *NoiseFactory) Local() cipher.PubKey {
	return nf.transportConfig.StaticPublic
}

// Type returns the Transport type.
func (nf *NoiseFactory) Type() string {
	return "noise"
}

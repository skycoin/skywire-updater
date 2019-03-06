package api

import (
	"context"
	"net/http"
	"net/rpc"
	"time"

	"github.com/skycoin/skywire-updater/pkg/update"
)

func handleRPC(g Gateway) http.Handler {
	rs := rpc.NewServer()
	if err := rs.RegisterName(rpcPrefix, &RPC{g: g}); err != nil {
		log.WithError(err).Fatalln("failed to register RPC")
		return nil
	}
	return rs
}

// RPC can be registered in a rpc.Server
type RPC struct {
	g Gateway
}

// Services lists services,
func (r *RPC) Services(_ *struct{}, services *[]string) error {
	*services = r.g.Services()
	return nil
}

// CheckIn is the input for Check.
type CheckIn struct {
	Service  string
	Deadline time.Time
}

// Check checks for updates for the given service.
func (r *RPC) Check(in *CheckIn, out *update.Release) error {
	ctx := context.Background()
	if !in.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, in.Deadline)
		defer cancel()
	}
	release, err := r.g.Check(ctx, in.Service)
	*out = *release
	return err
}

// UpdateIn is the input for Update.
type UpdateIn struct {
	Service   string
	ToVersion string
	Deadline  time.Time
}

// Update updates the given service.
func (r *RPC) Update(in *UpdateIn, ok *bool) (err error) {
	ctx := context.Background()
	if !in.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, in.Deadline)
		defer cancel()
	}
	*ok, err = r.g.Update(ctx, in.Service, in.ToVersion)
	return err
}

// RPCClient calls RPC.
type RPCClient struct {
	*rpc.Client
}

// DialRPC dials to a given skywire-updater RPC server of address.
func DialRPC(addr string) (*RPCClient, error) {
	rc, err := rpc.DialHTTPPath("tcp", addr, "/rpc")
	if err != nil {
		return nil, err
	}
	return &RPCClient{Client: rc}, nil
}

// Call calls with prefix.
func (rc *RPCClient) Call(method string, args, reply interface{}) error {
	return rc.Client.Call(rpcPrefix+"."+method, args, reply)
}

// Go gos with prefix.
func (rc *RPCClient) Go(method string, args, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return rc.Client.Go(rpcPrefix+"."+method, args, reply, done)
}

// Services calls Services.
func (rc *RPCClient) Services() ([]string, error) {
	var services []string
	err := rc.Call("Services", &struct{}{}, &services)
	return services, err
}

// Check calls Check.
func (rc *RPCClient) Check(srvName string, deadline time.Time) (update.Release, error) {
	var out update.Release
	err := rc.Call("Method", &CheckIn{Service: srvName, Deadline: deadline}, &out)
	return out, err
}

// Update calls Update.
func (rc *RPCClient) Update(srvName, toVersion string, deadline time.Time) (bool, error) {
	var ok bool
	err := rc.Call("Update", &UpdateIn{Service: srvName, ToVersion: toVersion, Deadline: deadline}, &ok)
	return ok, err
}

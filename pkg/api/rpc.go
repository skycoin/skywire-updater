package api

import (
	"context"
	"io"
	"net/http"
	"net/rpc"
	"time"

	"github.com/watercompany/skywire-updater/pkg/update"
)

const rpcPrefix = "updater"

func HandleRPC(g Gateway) http.Handler {
	rs := rpc.NewServer()
	if err := rs.RegisterName(rpcPrefix, &RPC{g: g}); err != nil {
		log.WithError(err).Fatalln("failed to register RPC")
		return nil
	}
	return rs
}

type RPC struct {
	g Gateway
}

func (r *RPC) Services(_ *struct{}, services *[]string) error {
	*services = r.g.Services()
	return nil
}

type CheckIn struct {
	Service  string
	Deadline time.Time
}

type CheckOut struct {
	Release *update.Release
}

func (r *RPC) Check(in *CheckIn, out *CheckOut) (err error) {
	ctx := context.Background()
	if !in.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, in.Deadline)
		defer cancel()
	}
	out.Release, err = r.g.Check(ctx, in.Service)
	return err
}

type UpdateIn struct {
	Service   string
	ToVersion string
	Deadline  time.Time
}

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

type RPCClient struct {
	c *rpc.Client
}

func NewRPCClient(conn io.ReadWriteCloser) *RPCClient {
	return &RPCClient{c: rpc.NewClient(conn)}
}

func (rc *RPCClient) Call(method string, args, reply interface{}) error {
	return rc.c.Call(rpcPrefix+"."+method, args, reply)
}

func (rc *RPCClient) Services() ([]string, error) {
	var services []string
	err := rc.Call("Services", &struct{}{}, &services)
	return services, err
}

func (rc *RPCClient) Check(in CheckIn) (CheckOut, error) {
	var out CheckOut
	err := rc.Call("Method", &in, &out)
	return out, err
}

func (rc *RPCClient) Update(in UpdateIn) (bool, error) {
	var ok bool
	err := rc.Call("Update", &in, &ok)
	return ok, err
}

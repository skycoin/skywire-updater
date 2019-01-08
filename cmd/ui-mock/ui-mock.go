package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/watercompany/skywire-updater/pkg/checker"
)

var logger = logging.MustGetLogger("api")
var port string

func main() {
	flag.StringVar(&port, "port", "8989", "port where to listen")
	err := NewServerGateway().Start(":" + port)
	if err != nil {
		logger.Fatal(err)
	}
}

// ServerGateway implements gateway interface for REST server
type ServerGateway struct {
	server *http.Server
}

// NewServerGateway returns a ServerGateway
func NewServerGateway() *ServerGateway {
	return &ServerGateway{}
}

// Start starts the REST server gateway
func (s *ServerGateway) Start(addrs string) error {
	l, err := net.Listen("tcp", addrs)
	if err != nil {
		return err
	}

	s.server = &http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/update", s.Update)
	s.server.Handler = mux

	return s.server.Serve(l)
}

// Stop closes the REST server gateway
func (s *ServerGateway) Stop() error {
	return s.server.Shutdown(context.Background())
}

// Update gets the service that needs to be updated and updates it
// URI: /update
func (s *ServerGateway) Update(w http.ResponseWriter, r *http.Request) {
	var msg checker.NotifyMsg
	defer r.Body.Close()
	b := r.Body
	sr, err := ioutil.ReadAll(b)
	if err != nil {
		logger.Error(err)
	}
	fmt.Println("body is: ", string(sr))
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info("update notification")
	logger.Info(msg)
}

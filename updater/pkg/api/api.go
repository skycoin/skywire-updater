package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/watercompany/skywire-services/updater/pkg/config"
	"github.com/watercompany/skywire-services/updater/pkg/supervisor"
)

var logger = logging.MustGetLogger("api")

// Gateway represents an API to communicate with the node
type Gateway interface {
	Start(string) error
	Stop() error
}

// HTTPResponse represents the http response struct
type HTTPResponse struct {
	Error *HTTPError  `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

// HTTPError is included in an HTTPResponse
type HTTPError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// RegisterMessage represents t
type RegisterMessage struct {
	NotifyURL      string `json:"notify-url"`
	CurrentVersion string `json:"current-version"`
	Repository     string `json:"repository"`
}

// NewHTTPErrorResponse returns an HTTPResponse with the Error field populated
func NewHTTPErrorResponse(code int, msg string) HTTPResponse {
	if msg == "" {
		msg = http.StatusText(code)
	}

	return HTTPResponse{
		Error: &HTTPError{
			Code:    code,
			Message: msg,
		},
	}
}

// ClientError is used for non-200 API responses
type ClientError struct {
	Status     string
	StatusCode int
	Message    string
}

// NewClientError creates a ClientError
func NewClientError(status string, statusCode int, message string) ClientError {
	return ClientError{
		Status:     status,
		StatusCode: statusCode,
		Message:    strings.TrimRight(message, "\n"),
	}
}

func (e ClientError) Error() string {
	return e.Message
}

// ServerGateway implements gateway interface for REST server
type ServerGateway struct {
	server  *http.Server
	starter *supervisor.Supervisor
}

// NewServerGateway returns a ServerGateway
func NewServerGateway(conf *config.Configuration) *ServerGateway {
	return &ServerGateway{
		starter: supervisor.New(conf),
	}
}

// Start starts the REST server gateway
func (s *ServerGateway) Start(addrs string) error {
	l, err := net.Listen("tcp", addrs)
	if err != nil {
		return err
	}

	s.server = &http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/check/", s.Check)
	mux.HandleFunc("/update/", s.Update)
	mux.HandleFunc("/register/", s.Register)
	mux.HandleFunc("/unregister/", s.Unregister)
	s.server.Handler = mux

	return s.server.Serve(l)
}

// Stop closes the REST server gateway
func (s *ServerGateway) Stop() error {
	return s.server.Shutdown(context.Background())
}

// Check checks if there is an update for the given service
// URI: /check/:service_name
// Method: GET
func (s *ServerGateway) Check(w http.ResponseWriter, r *http.Request) {
	service := retrieveServiceFromURL(r.URL)
	err := s.starter.Check(service)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			NewHTTPErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, HTTPResponse{Data: service + " has update"})
}

// Update gets the service that needs to be updated and updates it
// URI: /update/:service_name
// Method: POST
func (s *ServerGateway) Update(w http.ResponseWriter, r *http.Request) {
	service := retrieveServiceFromURL(r.URL)
	err := s.starter.Update(service)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			NewHTTPErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, HTTPResponse{Data: service + " updated"})
}

// Register registers a new service into updater in order to look for new versions of it.
// URI: /register/:service_name
// Method: POST
// Content-Type: application/json
// Body: {
// 	"notify-url":"<url where to send a POST request upon new version available>",
// 	"current-version":"<current version of the service to register>",
// 	"repository":"<repository where to check for updates>"
// }
func (s *ServerGateway) Register(w http.ResponseWriter, r *http.Request) {
	service := retrieveServiceFromURL(r.URL)
	var registerMsg RegisterMessage

	err := json.NewDecoder(r.Body).Decode(&registerMsg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			NewHTTPErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	s.starter.Register(service, registerMsg.Repository,
		registerMsg.NotifyURL, registerMsg.CurrentVersion)
}

// Unregister unregisters a service from updater, which will stop looking for new versions
// URI: /register/:service_name
// Method: POST
func (s *ServerGateway) Unregister(w http.ResponseWriter, r *http.Request) {
	service := retrieveServiceFromURL(r.URL)

	err := s.starter.Unregister(service)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError,
			NewHTTPErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, HTTPResponse{Data: service + " unregistered"})
}

// retrievePkFromURL returns the id used on endpoints of the form path/:pk
// it doesn't checks if the endpoint has this form and can fail with other
// endpoint forms
func retrieveServiceFromURL(url *url.URL) string {
	splittedPath := strings.Split(url.EscapedPath(), "/")
	return splittedPath[len(splittedPath)-1]
}

// writeJSON writes a json object on a http.ResponseWriter with the given code,
// panics on marshaling error
func writeJSON(w http.ResponseWriter, code int, object interface{}) {
	jsonObject, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(jsonObject)
	if err != nil {
		logger.Error(err)
	}
}

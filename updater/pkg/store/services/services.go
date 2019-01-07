package services

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/skycoin/skycoin/src/util/file"
)

var storeSingleton Storer

// Storer represents an object that can store a service configuration as well as retrieve that
// configuration
type Storer interface {
	Get(string) *Service
	Store(*Service)
}

// InitStorer starts a storer object of the given store kind
func InitStorer(store string) {
	switch store {
	case "json":
		storePath := filepath.Join(file.UserHome(), ".updater", "services.json")
		storeSingleton = newJSONStore(storePath)
	}
}

// GetStore returns a singleton initialized storer object
func GetStore() Storer {
	return storeSingleton
}

// Services represents a map of service updates
type Services map[string]*Service

func (s Services) get(name string) *Service {
	if _, ok := s[name]; !ok {
		oldTime := time.Date(1999, 10, 1, 1, 1, 1, 1, time.UTC)
		service := &Service{
			Name:        name,
			LastUpdated: NewTimeJSON(oldTime),
		}
		s[name] = service
	}
	return s[name]
}

func (s Services) set(service *Service) {
	s[service.Name] = service
}

// Service represents a service related configuration which holds the last update information
type Service struct {
	Name        string    `json:"name"`
	LastUpdated *TimeJSON `json:"last_updated"`
}

// MarshalJSON implements marshaler
func (s *Service) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		LastUpdated TimeJSON `json:"last_updated"`
		Name        string   `json:"name"`
	}{
		*s.LastUpdated,
		s.Name,
	})
}

// TimeJSON is a custom json time type to marshal and unmarshal golang stdylib time
type TimeJSON struct {
	time.Time
}

// NewTimeJSON returns a TimeJSON of the given time
func NewTimeJSON(t time.Time) *TimeJSON {
	return &TimeJSON{
		Time: t,
	}
}

// UnmarshalJSON implements unmarshaler
func (t *TimeJSON) UnmarshalJSON(b []byte) error {
	// On the first and last character on the array we have an additional
	// double quote characters, which will cause an error on parse
	cleanInput := b[1 : len(b)-1]
	parsedTime, err := time.Parse(time.RFC3339, string(cleanInput))
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

// MarshalJSON implements marshaler
func (t *TimeJSON) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(time.RFC3339)), nil
}

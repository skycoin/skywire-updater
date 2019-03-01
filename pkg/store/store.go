package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/skycoin/skycoin/src/util/logging"
)

// Update represents an update entry.
type Update struct {
	Tag       string `json:"tag,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// IsEmpty checks whether the update is empty.
func (u Update) IsEmpty() bool {
	return u.Tag == "" && u.Timestamp == 0
}

// Store represents a database implementation.
type Store interface {
	ServiceLastUpdate(srvName string) Update
	SetServiceLastUpdate(srvName string, last Update)
	Close() error
}

// JSON implements Store.
type JSON struct {
	*os.File
	data map[string]Update // key: srvName, value: Update
	mu   sync.RWMutex
	log  *logging.Logger
}

// NewJSON creates a new JSON Store implementation.
func NewJSON(filePath string) (*JSON, error) {
	db := &JSON{
		data: make(map[string]Update),
		log:  logging.MustGetLogger("store(JSON)"),
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0744); err != nil {
		return nil, fmt.Errorf("failed to create db file: %s", err.Error())
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to recreate state: %s", err.Error())
	}
	db.File = f

	if err := json.NewDecoder(f).Decode(&db.data); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read '%s': %s", filePath, err.Error())
	}

	return db, nil
}

// ServiceLastUpdate obtains the last update for a given service..
func (j *JSON) ServiceLastUpdate(srvName string) Update {
	j.mu.RLock()
	defer j.mu.RUnlock()

	update, ok := j.data[srvName]
	j.log.Infof("data[%s]: (%v) %v", srvName, ok, update)
	if !ok {
		return Update{}
	}
	return update
}

// SetServiceLastUpdate sets a last update for a given service.
func (j *JSON) SetServiceLastUpdate(srvName string, last Update) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.data[srvName] = last
	if err := j.Truncate(0); err != nil {
		j.log.WithError(err).Fatal()
	}
	if _, err := j.Seek(0, 0); err != nil {
		j.log.WithError(err).Fatal()
	}
	if err := json.NewEncoder(j).Encode(&j.data); err != nil {
		j.log.WithError(err).Fatal()
	}
}

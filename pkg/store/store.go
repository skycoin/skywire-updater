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

type Update struct {
	Tag       string `json:"tag,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

func (u Update) IsEmpty() bool {
	return u.Tag == "" && u.Timestamp == 0
}

type Store interface {
	ServiceLastUpdate(srvName string) Update
	SetServiceLastUpdate(srvName string, last Update)
	Close() error
}

type JSON struct {
	*os.File
	data map[string]Update // key: srvName, value: Update
	mu   sync.RWMutex
	log  *logging.Logger
}

func NewJSON(filePath string) (*JSON, error) {
	db := &JSON{
		data: make(map[string]Update),
		log:  logging.MustGetLogger("store.JSON"),
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

func (j *JSON) SetServiceLastUpdate(srvName string, last Update) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.data[srvName] = last
	if err := j.Truncate(0); err != nil {
		j.log.WithError(err).Fatal()
		return
	}
	if _, err := j.Seek(0, 0); err != nil {
		j.log.WithError(err).Fatal()
		return
	}
	if err := json.NewEncoder(j).Encode(&j.data); err != nil {
		j.log.WithError(err).Fatal()
		return
	}
	return
}

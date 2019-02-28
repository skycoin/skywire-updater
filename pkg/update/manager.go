package update

import (
	"context"
	"errors"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/watercompany/skywire-updater/pkg/store"
)

var (
	// ErrServiceNotFound ...
	ErrServiceNotFound = errors.New("service of given name is not found")
)

type srvEntry struct {
	ServiceConfig
	Checker
	Updater
	sync.Mutex
}

// Manager ...
type Manager struct {
	services map[string]srvEntry
	mu       sync.RWMutex
	db       store.Store
	log      *logging.Logger
}

// NewManager ...
func NewManager(db store.Store, scriptsDir string, conf *Config) *Manager {
	d := &Manager{
		services: make(map[string]srvEntry),
		db:       db,
		log:      logging.MustGetLogger("daemon"),
	}
	for srvName := range conf.Services {
		srvConf := *conf.Services[srvName]
		srvConf.Checker.Script = path.Join(scriptsDir, srvConf.Checker.Script)
		srvConf.Updater.Script = path.Join(scriptsDir, srvConf.Updater.Script)
		d.services[srvName] = srvEntry{
			ServiceConfig: srvConf,
			Checker:       NewChecker(logging.MustGetLogger(srvName+".checker"), db, srvName, srvConf),
			Updater:       NewUpdater(logging.MustGetLogger(srvName+".updater"), srvName, srvConf),
		}
	}
	return d
}

// Services ...
func (d *Manager) Services() []string {
	d.mu.RLock()
	var srvNames []string
	for srvName := range d.services {
		srvNames = append(srvNames, srvName)
	}
	sort.Strings(srvNames)
	d.mu.RUnlock()
	return srvNames
}

// Check ...
func (d *Manager) Check(ctx context.Context, srvName string) (*Release, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	srv, ok := d.services[srvName]
	if !ok {
		return nil, ErrServiceNotFound
	}

	srv.Lock()
	release, err := srv.Check(ctx)
	srv.Unlock()

	return release, err
}

// Update ...
func (d *Manager) Update(ctx context.Context, srvName, toVersion string) (bool, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	srv, ok := d.services[srvName]
	if !ok {
		return false, ErrServiceNotFound
	}

	srv.Lock()
	defer srv.Unlock()

	updated, err := srv.Update(ctx, toVersion)
	if err != nil {
		return false, err
	}
	if updated {
		entry := store.Update{
			Tag:       toVersion,
			Timestamp: time.Now().UnixNano(),
		}
		d.db.SetServiceLastUpdate(srvName, entry)
	}
	return updated, nil
}

// Close ...
func (d *Manager) Close() error {
	d.mu.Lock()
	err := d.db.Close()
	d.mu.Unlock()

	return err
}

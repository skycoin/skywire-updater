package update

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/skywire-updater/pkg/store"
)

var (
	// ErrServiceNotFound occurs when service is not found.
	ErrServiceNotFound = errors.New("service of given name is not found")

	log = logging.MustGetLogger("skywire-updater")
)

type srvEntry struct {
	ServiceConfig
	Checker
	Updater
	sync.Mutex
}

// Manager manages checkers and updaters for services.
type Manager struct {
	global   ServiceDefaultsConfig
	services map[string]srvEntry
	mu       sync.RWMutex
	db       store.Store
}

// NewManager creates a new manager.
func NewManager(db store.Store, conf *Config) *Manager {
	d := &Manager{
		global:   conf.Services.Defaults,
		services: make(map[string]srvEntry),
		db:       db,
	}
	for name, srv := range conf.Services.Services {
		d.services[name] = srvEntry{
			ServiceConfig: *srv,
			Checker:       NewChecker(db, name, *srv, &d.global),
			Updater:       NewUpdater(name, *srv, &d.global),
		}
	}
	return d
}

// Services lists the available services.
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

// Check checks for updates for a given service.
func (d *Manager) Check(ctx context.Context, srvName string) (*Release, error) {
	d.mu.RLock()
	srv, ok := d.services[srvName]
	d.mu.RUnlock()
	if !ok {
		return nil, ErrServiceNotFound
	}
	srv.Lock()
	release, err := srv.Check(ctx)
	srv.Unlock()
	return release, err
}

// Update updates given service to provided version.
func (d *Manager) Update(ctx context.Context, srvName, toVersion string) (bool, error) {
	d.mu.RLock()
	srv, ok := d.services[srvName]
	d.mu.RUnlock()
	if !ok {
		return false, ErrServiceNotFound
	}
	srv.Lock()
	updated, err := srv.Update(ctx, toVersion)
	srv.Unlock()
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

// Close closes the manager.
func (d *Manager) Close() error {
	d.mu.Lock()
	d.services = make(map[string]srvEntry)
	d.mu.Unlock()
	return d.db.Close()
}

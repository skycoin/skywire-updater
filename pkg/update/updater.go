package update

import (
	"context"
	"os/exec"

	"github.com/skycoin/skycoin/src/util/logging"
)

// UpdaterType determines the updater type.
type UpdaterType string

const (
	// ScriptUpdaterType represents the script updater type.
	ScriptUpdaterType = UpdaterType("script")
)

var updaterTypes = []UpdaterType{
	ScriptUpdaterType,
}

// Updater updates a given service.
type Updater interface {
	Update(ctx context.Context, toVersion string) (bool, error)
}

// NewUpdater creates a new updater.
func NewUpdater(srvName string, c ServiceConfig, d *ServiceDefaultsConfig) Updater {
	switch c.Updater.Type {
	case ScriptUpdaterType:
		return NewScriptUpdater(srvName, c, d)
	default:
		log.Fatalf("invalid updater type '%s' at 'services[%s].updater.type' when expecting: %v",
			c.Updater.Type, updaterTypes)
		return nil
	}
}

// ScriptUpdater is an implementation of updater using scripts.
type ScriptUpdater struct {
	c   ServiceConfig
	d   *ServiceDefaultsConfig
	log *logging.Logger
}

// NewScriptUpdater creates a new ScriptUpdater.
func NewScriptUpdater(srvName string, c ServiceConfig, d *ServiceDefaultsConfig) *ScriptUpdater {
	return &ScriptUpdater{
		c:   c,
		d:   d,
		log: logging.MustGetLogger("script-updater." + srvName),
	}
}

// Update updates the given service to specified version.
func (cu *ScriptUpdater) Update(ctx context.Context, version string) (bool, error) {
	update := cu.c.Updater
	cmd := exec.Command(update.Interpreter, append([]string{update.Script}, update.Args...)...) //nolint:gosec
	cmd.Env = UpdaterEnvs(cu.d, &cu.c, version)

	return ExecuteScript(ctx, cu.log, cmd)
}

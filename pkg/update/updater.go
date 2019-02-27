package update

import (
	"context"
	"os/exec"

	"github.com/skycoin/skycoin/src/util/logging"
)

// UpdaterType ...
type UpdaterType string

const (
	// ScriptUpdaterType ...
	ScriptUpdaterType = UpdaterType("script")
)

var updaterTypes = []UpdaterType{
	ScriptUpdaterType,
}

// Updater ...
type Updater interface {
	Update(ctx context.Context, toVersion string) (bool, error)
}

// NewUpdater ...
func NewUpdater(log *logging.Logger, srvName string, srvConfig ServiceConfig) Updater {
	switch srvConfig.Updater.Type {
	case ScriptUpdaterType:
		return NewScriptUpdater(log, srvName, srvConfig)
	default:
		log.Fatalf("invalid updater type '%s' at 'services[%s].updater.type' when expecting: %v",
			srvConfig.Updater.Type, srvName, updaterTypes)
		return nil
	}
}

// ScriptUpdater ...
type ScriptUpdater struct {
	srvName string
	c       ServiceConfig
	log     *logging.Logger
}

// NewScriptUpdater ...
func NewScriptUpdater(log *logging.Logger, srvName string, c ServiceConfig) *ScriptUpdater {
	return &ScriptUpdater{
		srvName: srvName,
		c:       c,
		log:     log,
	}
}

// Update ...
func (cu *ScriptUpdater) Update(ctx context.Context, version string) (bool, error) {
	update := cu.c.Updater
	cmd := exec.Command(update.Interpreter, append([]string{update.Script}, update.Args...)...) //nolint:gosec
	cmd.Env = append(cu.c.UpdaterEnvs(), cmdEnv(EnvToVersion, version))

	return executeScript(ctx, cmd, cu.log)
}

package update

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/stretchr/testify/assert"
)

func TestCheckerEnvs(t *testing.T) {
	type Test struct {
		Default DefaultsConfig
		Service ServiceConfig
		Script  string
	}
	do := func(t *testing.T, c Test) {
		fName, rm := prepareScript(t, c.Script)
		defer rm()
		cmd := exec.Command("/bin/bash", fName)
		cmd.Env = CheckerEnvs(&c.Default, &c.Service)
		ok, err := ExecuteScript(context.TODO(), logging.MustGetLogger("test"), cmd)
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	t.Run(fmt.Sprintf("default(%s)", EnvMainBranch), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvMainBranch, "master")},
			},
			Service: ServiceConfig{},
			Script:  fmt.Sprintf(`if [ "${%s}" != "master" ]; then exit 1; fi`, EnvMainBranch),
		})
	})
	t.Run(fmt.Sprintf("service_override(%s)", EnvMainBranch), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvMainBranch, "master")},
			},
			Service: ServiceConfig{
				MainBranch: "stable",
			},
			Script: fmt.Sprintf(`if [ "${%s}" != "stable" ]; then exit 1; fi`, EnvMainBranch),
		})
	})
	t.Run(fmt.Sprintf("checker_override_1(%s)", EnvMainBranch), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvMainBranch, "master")},
			},
			Service: ServiceConfig{
				MainBranch: "stable",
				Checker: CheckerConfig{
					Envs: []string{MakeEnv(EnvMainBranch, "develop")},
				},
			},
			Script: fmt.Sprintf(`if [ "${%s}" != "develop" ]; then exit 1; fi`, EnvMainBranch),
		})
	})
	t.Run(fmt.Sprintf("checker_override_2(%s)", EnvMainBranch), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvMainBranch, "master")},
			},
			Service: ServiceConfig{
				Checker: CheckerConfig{
					Envs: []string{MakeEnv(EnvMainBranch, "develop")},
				},
			},
			Script: fmt.Sprintf(`if [ "${%s}" != "develop" ]; then exit 1; fi`, EnvMainBranch),
		})
	})
	t.Run("default(NEW_ENV)", func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{"NEW_ENV=env_value"},
			},
			Service: ServiceConfig{
				Checker: CheckerConfig{
					Envs: []string{"ANOTHER_ENV=another_value"},
				},
			},
			Script: `if [ "${NEW_ENV}" != "env_value" ] || [ "${ANOTHER_ENV}" != "another_value" ]; then exit 1; fi`,
		})
	})
	t.Run("checker_override(NEW_ENV)", func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{"NEW_ENV=env_value"},
			},
			Service: ServiceConfig{
				Checker: CheckerConfig{
					Envs: []string{"NEW_ENV=changed_value", "ANOTHER_ENV=another_value"},
				},
			},
			Script: `if [ "${NEW_ENV}" != "changed_value" ] || [ "${ANOTHER_ENV}" != "another_value" ]; then exit 1; fi`,
		})
	})
}

func TestUpdaterEnvs(t *testing.T) {
	type Test struct {
		Default   DefaultsConfig
		Service   ServiceConfig
		ToVersion string
		Script    string
	}
	do := func(t *testing.T, c Test) {
		fName, rm := prepareScript(t, c.Script)
		defer rm()
		cmd := exec.Command("/bin/bash", fName)
		cmd.Env = UpdaterEnvs(&c.Default, &c.Service, c.ToVersion)
		ok, err := ExecuteScript(context.TODO(), logging.MustGetLogger("test"), cmd)
		assert.NoError(t, err)
		assert.True(t, ok)
	}
	t.Run(fmt.Sprintf("default(%s)", EnvToVersion), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvToVersion, "v1.0")},
			},
			Script: fmt.Sprintf(`if [ "${%s}" != "v1.0" ]; then exit 1; fi`, EnvToVersion),
		})
	})
	t.Run(fmt.Sprintf("service(%s)", EnvToVersion), func(t *testing.T) {
		do(t, Test{
			Service: ServiceConfig{
				Updater: UpdaterConfig{
					Envs: []string{MakeEnv(EnvToVersion, "v1.1")},
				},
			},
			Script: fmt.Sprintf(`if [ "${%s}" != "v1.1" ]; then exit 1; fi`, EnvToVersion),
		})
	})
	t.Run(fmt.Sprintf("updater_override(%s)", EnvToVersion), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvToVersion, "v1.0")},
			},
			Service: ServiceConfig{
				Updater: UpdaterConfig{
					Envs: []string{MakeEnv(EnvToVersion, "v2.0")},
				},
			},
			Script: fmt.Sprintf(`if [ "${%s}" != "v2.0" ]; then exit 1; fi`, EnvToVersion),
		})
	})
	t.Run(fmt.Sprintf("updater_update_override_1(%s)", EnvToVersion), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvToVersion, "v1.0")},
			},
			Service: ServiceConfig{
				Updater: UpdaterConfig{
					Envs: []string{MakeEnv(EnvToVersion, "v2.0")},
				},
			},
			ToVersion: "v3.0",
			Script:    fmt.Sprintf(`if [ "${%s}" != "v3.0" ]; then exit 1; fi`, EnvToVersion),
		})
	})
	t.Run(fmt.Sprintf("updater_update_override_2(%s)", EnvToVersion), func(t *testing.T) {
		do(t, Test{
			Default: DefaultsConfig{
				Envs: []string{MakeEnv(EnvToVersion, "v1.0")},
			},
			ToVersion: "v3.0",
			Script:    fmt.Sprintf(`if [ "${%s}" != "v3.0" ]; then exit 1; fi`, EnvToVersion),
		})
	})
	t.Run(fmt.Sprintf("updater_update_override_3(%s)", EnvToVersion), func(t *testing.T) {
		do(t, Test{
			ToVersion: "v3.0",
			Script:    fmt.Sprintf(`if [ "${%s}" != "v3.0" ]; then exit 1; fi`, EnvToVersion),
		})
	})
}

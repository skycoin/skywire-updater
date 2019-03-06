package update

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testUpdateScript = `#!/bin/bash

echo "repo: ${SKYUPD_REPO}"
echo "branch: ${SKYUPD_MAIN_BRANCH}"
echo "process: ${SKYUPD_MAIN_PROCESS}"
echo "version: ${SKYUPD_TO_VERSION}"
shift 4

echo "args: ${@}"
`

func TestScriptUpdater_Update(t *testing.T) {
	fName, rm := prepareScript(t, testUpdateScript)
	defer rm()

	c := ServiceConfig{
		Repo:        "domain.com/org/repo",
		MainBranch:  "branch",
		MainProcess: "run",
		Updater: UpdaterConfig{
			Type:        ScriptUpdaterType,
			Interpreter: "/bin/bash",
			Script:      fName,
			Args:        []string{"arg1"},
		},
	}
	updater := NewUpdater("my-service", c, new(ServiceDefaultsConfig))

	ok, err := updater.Update(context.TODO(), "v1.0")
	assert.NoError(t, err)
	assert.True(t, ok)
}

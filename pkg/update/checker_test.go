package update

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/stretchr/testify/require"

	"github.com/watercompany/skywire-updater/pkg/store"
)

const testCheckScript = `#!/bin/bash

echo "repo: ${SKYUPD_REPO}"
echo "branch: ${SKYUPD_MAIN_BRANCH}"
echo "process: ${SKYUPD_MAIN_PROCESS}"
shift 3

echo "args: ${@}"
`

func TestScriptChecker_Check(t *testing.T) {
	fName, rm := prepareScript(t, testCheckScript)
	defer rm()

	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	defer func() {
		require.NoError(t, os.Remove(f.Name()))
	}()
	j, err := store.NewJSON(f.Name())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, j.Close())
	}()

	c := ServiceConfig{
		Repo:        "domain.com/org/repo",
		MainBranch:  "branch",
		MainProcess: "run",
		Checker: CheckerConfig{
			Type:        ScriptCheckerType,
			Interpreter: "/bin/bash",
			Script:      fName,
			Args:        []string{"arg1"},
		},
	}
	checker := NewChecker(logging.MustGetLogger("my_service"), j, "my_service", c, new(DefaultConfig))

	r, err := checker.Check(context.TODO())
	require.NoError(t, err)
	require.True(t, r.HasUpdate)
	require.Equal(t, c.Checker.Type, r.CheckerType)
	t.Log(r)
}

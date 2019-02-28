package update

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testScript = `
#!/bin/bash

echo "repo: ${SKYUPD_REPO}"
echo "version: ${SKYUPD_TO_VERSION}"
shift 2

echo "args: $@"
`

func TestCustomUpdater_Update(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(testScript)
	require.NoError(t, err)

	t.Run("ScriptUpdater", func(t *testing.T) {
		c := ServiceConfig{
			Repo: "domain.com/org/repo",
			Updater: UpdaterConfig{
				Type:        ScriptUpdaterType,
				Interpreter: "/bin/bash",
				Script:      f.Name(),
				Args:        []string{"arg1"},
			},
		}
		updater := NewUpdater(logging.MustGetLogger("my_service"), "my_service", c)

		ok, err := updater.Update(context.TODO(), "v1.0")
		assert.NoError(t, err)
		assert.True(t, ok)
	})
}

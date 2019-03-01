package update

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareScript(t *testing.T, script string) (string, func()) {
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	_, err = f.WriteString(script)
	require.NoError(t, err)

	require.NoError(t, f.Close())
	rm := func() {
		require.NoError(t, os.Remove(f.Name()))
	}
	return f.Name(), rm
}

func TestExecuteScript(t *testing.T) {
	t.Run("exit_code_0", func(t *testing.T) {
		fName, rm := prepareScript(t, "exit 0")
		defer rm()

		ok, err := ExecuteScript(context.Background(),
			logging.MustGetLogger("0"),
			exec.Command("/bin/bash", fName))

		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("exit_code_1", func(t *testing.T) {
		fName, rm := prepareScript(t, "exit 1")
		defer rm()

		ok, err := ExecuteScript(context.Background(),
			logging.MustGetLogger("1"),
			exec.Command("/bin/bash", fName))

		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("exit_code_other", func(t *testing.T) {
		fName, rm := prepareScript(t, "exit 2")
		defer rm()

		ok, err := ExecuteScript(context.Background(),
			logging.MustGetLogger("2"),
			exec.Command("/bin/bash", fName))

		assert.Error(t, err)
		assert.False(t, ok)
	})
}

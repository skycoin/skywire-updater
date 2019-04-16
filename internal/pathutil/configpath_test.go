package pathutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var wd, err = os.Getwd()

var defaultDir = ConfigPaths{
	WorkingDirLoc: filepath.Join(wd, "config.yml"),
	HomeLoc:       filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml"),
	LocalLoc:      "/usr/local/skycoin/skywire-updater/config.yml",
}

type inputs = struct {
	args      []string
	argsIndex int
	env       string
	defaults  ConfigPaths
}

var tables = []struct {
	description string
	in          inputs
	expectedMsg string
}{
	{
		description: "defaultConfig",
		in:          inputs{[]string{}, 0, "SW_MANAGER_CONFIG", defaultDir},
		expectedMsg: filepath.Join(HomeDir(), ".skycoin/skywire-updater/config.yml"),
	},
}

func TestFindConfigPath(t *testing.T) {

	for _, tt := range tables {
		t.Run(tt.description, func(t *testing.T) {

			assert := assert.New(t)

			// Assert err from os.Getwd()
			assert.Nil(err)

			configPath := FindConfigPath(tt.in.args, tt.in.argsIndex, tt.in.env, tt.in.defaults)

			assert.Equal(configPath, tt.expectedMsg, "This should be equal!")
		})
	}
}
